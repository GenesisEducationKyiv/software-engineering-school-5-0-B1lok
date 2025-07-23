//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"

	"testing"
)

func TestSuccessfulSubscription(t *testing.T) {

	pw, err := playwright.Run()

	if err != nil {
		log.Fatal().Err(err).Msg("could not start Playwright")
	}
	defer func() {
		if err := pw.Stop(); err != nil {
			log.Fatal().Err(err).Msg("could not stop Playwright")
		}
	}()

	for _, browserData := range getBrowsers(pw) {
		t.Run(fmt.Sprintf("Browser_%s", browserData.name), func(t *testing.T) {
			browser, err := browserData.browser.Launch(playwright.BrowserTypeLaunchOptions{
				Headless: playwright.Bool(true),
			})
			if err != nil {
				t.Fatalf("could not launch browser %s: %v", browserData.name, err)
			}
			defer func() {
				if err := browser.Close(); err != nil {
					log.Fatal().Err(err).Msg("could not close browser")
				}
			}()

			page, err := browser.NewPage()
			if err != nil {
				t.Fatalf("could not create page: %v", err)
			}
			defer func() {
				if err := browser.Close(); err != nil {
					log.Fatal().Err(err).Msg("could not close browser")
				}
			}()

			_, err = page.Goto("http://localhost:8080")
			if err != nil {
				t.Fatalf("could not navigate: %v", err)
			}

			err = page.Fill("#city", "Kyiv")
			if err != nil {
				t.Fatalf("could not fill city: %v", err)
			}

			err = page.Fill("#email", generateRandomEmail())
			if err != nil {
				t.Fatalf("could not fill email: %v", err)
			}

			_, err = page.SelectOption("#frequency", playwright.SelectOptionValues{
				Values: &[]string{"daily"},
			})
			if err != nil {
				t.Fatalf("could not select frequency: %v", err)
			}

			err = page.Click("#submit-button")
			if err != nil {
				t.Fatalf("could not click submit button: %v", err)
			}

			_, err = page.WaitForSelector("#notification", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(10000),
			})
			if err != nil {
				t.Fatalf("notification did not appear: %v", err)
			}

			notificationText, err := page.InnerText("#notification")
			if err != nil {
				t.Fatalf("could not retrieve notification text: %v", err)
			}

			isSuccess, _ := page.IsVisible("#notification.success")
			isError, _ := page.IsVisible("#notification.error")

			if isSuccess {
				fmt.Printf("%s browser test - SUCCESS: %s\n", browserData.name, notificationText)
			} else if isError {
				t.Errorf("%s browser test - ERROR: %s\n", browserData.name, notificationText)
			} else {
				t.Errorf("Unexpected notification state for %s: %s", browserData.name, notificationText)
			}
		})
	}
}

func TestFormValidationDetailed(t *testing.T) {
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start Playwright: %v", err)
	}
	defer func() {
		if err := pw.Stop(); err != nil {
			log.Fatal().Err(err).Msg("could not stop Playwright")
		}
	}()

	for _, browserData := range getBrowsers(pw) {
		t.Run(fmt.Sprintf("Browser_%s", browserData.name), func(t *testing.T) {
			browser, err := browserData.browser.Launch(playwright.BrowserTypeLaunchOptions{
				Headless: playwright.Bool(true),
			})
			if err != nil {
				t.Fatalf("could not launch browser: %v", err)
			}
			defer func() {
				if err := browser.Close(); err != nil {
					log.Fatal().Err(err).Msg("could not close browser")
				}
			}()

			page, err := browser.NewPage()
			if err != nil {
				t.Fatalf("could not create page: %v", err)
			}
			defer func() {
				if err := browser.Close(); err != nil {
					log.Fatal().Err(err).Msg("could not close browser")
				}
			}()

			_, err = page.Goto("http://localhost:8080")
			if err != nil {
				t.Fatalf("could not navigate: %v", err)
			}
			err = page.Click("#submit-button")
			if err != nil {
				t.Fatalf("could not click submit: %v", err)
			}

			cityErrorVisible, _ := page.IsVisible("#city-error.visible")
			emailErrorVisible, _ := page.IsVisible("#email-error.visible")
			frequencyErrorVisible, _ := page.IsVisible("#frequency-error.visible")

			if !cityErrorVisible {
				t.Error("City error should be visible")
			}
			if !emailErrorVisible {
				t.Error("Email error should be visible")
			}
			if !frequencyErrorVisible {
				t.Error("Frequency error should be visible")
			}
		})
	}
}

func TestInvalidCityInput(t *testing.T) {

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("could not start Playwright")
	}
	defer func() {
		if err := pw.Stop(); err != nil {
			log.Fatal().Err(err).Msg("could not stop Playwright")
		}
	}()

	for _, browserData := range getBrowsers(pw) {
		t.Run(fmt.Sprintf("Browser_%s", browserData.name), func(t *testing.T) {
			browser, err := browserData.browser.Launch(playwright.BrowserTypeLaunchOptions{
				Headless: playwright.Bool(true),
			})
			if err != nil {
				t.Fatalf("could not launch browser %s: %v", browserData.name, err)
			}
			defer func() {
				if err := browser.Close(); err != nil {
					log.Fatal().Err(err).Msg("could not close browser")
				}
			}()

			page, err := browser.NewPage()
			if err != nil {
				t.Fatalf("could not create page: %v", err)
			}
			defer func() {
				if err := browser.Close(); err != nil {
					log.Fatal().Err(err).Msg("could not close browser")
				}
			}()

			_, err = page.Goto("http://localhost:8080")
			if err != nil {
				t.Fatalf("could not navigate: %v", err)
			}

			err = page.Fill("#city", "InvalidCity")
			if err != nil {
				t.Fatalf("could not fill city: %v", err)
			}

			err = page.Fill("#email", generateRandomEmail())
			if err != nil {
				t.Fatalf("could not fill email: %v", err)
			}

			_, err = page.SelectOption("#frequency", playwright.SelectOptionValues{
				Values: &[]string{"daily"},
			})
			if err != nil {
				t.Fatalf("could not select frequency: %v", err)
			}

			err = page.Click("#submit-button")
			if err != nil {
				t.Fatalf("could not click submit button: %v", err)
			}

			_, err = page.WaitForSelector("#notification", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(10000),
			})
			if err != nil {
				t.Fatalf("notification did not appear: %v", err)
			}

			notificationText, err := page.InnerText("#notification")
			if err != nil {
				t.Fatalf("could not retrieve notification text: %v", err)
			}

			isError, _ := page.IsVisible("#notification.error")

			if !isError || notificationText != "Please check your information and try again." {
				t.Errorf("%s browser test - ERROR: %s\n", browserData.name, notificationText)
			}
		})
	}
}

func getBrowsers(pw *playwright.Playwright) []struct {
	name    string
	browser playwright.BrowserType
} {
	return []struct {
		name    string
		browser playwright.BrowserType
	}{
		{"chromium", pw.Chromium},
		{"firefox", pw.Firefox},
		{"webkit", pw.WebKit},
	}
}

func generateRandomEmail() string {
	return fmt.Sprintf("user%s@example.com", uuid.New().String())
}
