package eurid

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/caarlos0/domain_exporter/internal/client"
)

var (
	format = "02 Jan 2006"
	expiryRE = regexp.MustCompile(`>Registered<\/div>\n *<div class="stat-value">(.+)<`)
)

type euridClient struct{}

// NewClient returns a new EURid client.
func NewClient() client.Client {
	return euridClient{}
}

func (euridClient) ExpireTime(ctx context.Context, domain string) (time.Time, error) {
	log.Debug().Msgf("trying eurid client for %s", domain)
	s := strings.Split(domain, ".")
	tld := s[len(s) - 1]
	if tld != "eu" && tld != "ею" && tld != "ευ" {
		return time.Time{}, fmt.Errorf("unsupported tld: %s", tld)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://whois.eurid.eu/en/search/", nil)
	q := req.URL.Query()
	q.Add("domain", domain)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to do eurid request: %w", err)
	}
	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, err
	}
	result := expiryRE.FindStringSubmatch(string(rb))
	if len(result) < 2 {
		return time.Time{}, fmt.Errorf("failed to parse eurid response: %w", err)
	}
	regDate, err := time.Parse(format, result[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date: %w", err)
	}
	// Expiry date is calculated from registered day
	// https://eurid.eu/en/register-a-eu-domain/rules-for-eu-domains/
	date, currDate := regDate.AddDate(1, 0, 0), time.Now()
	for date.Before(currDate) {
		date = date.AddDate(1, 0, 0)
	}
	log.Debug().Msgf("domain %q will expire at %q", domain, date.String())
	return date, nil
}
