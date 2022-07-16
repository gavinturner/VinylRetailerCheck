package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/cmd"
	"github.com/gavinturner/vinylretailers/db"
	"github.com/gavinturner/vinylretailers/retailers"
	"github.com/gavinturner/vinylretailers/util/email"
	"github.com/gavinturner/vinylretailers/util/log"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"time"
)

const (
	// delay for ten minutes between attepts to produce reports
	STARTUP_DELAY_SECS                 = 10
	DBSTARTUP_TIMEOUT_SECS             = 30
	DELAY_BETWEEN_REPORTING_POLLS_SECS = 10
)

//
// reporter.main()
// Represents the process body of the reporter pod. ...
// @see scheduler.main()
// @see scanner.main()
//
func main() {

	// sleep on startup to let the infra pods get started up
	time.Sleep(time.Duration(STARTUP_DELAY_SECS) * time.Second)

	// use the database config and initialise a postgres connection (will panic if incomplete)
	psqlDB, err := cmd.InitialiseDbConnection()
	if err != nil {
		panic(err)
	}
	if psqlDB == nil {
		panic("db pointer is null?")
	}
	defer psqlDB.Close()
	vinylDS := db.NewDB(psqlDB)

	// make sure that the db is up. keep trying every second until it is
	log.Debugf("Retail vinyl reporter starts..\n")
	err = vinylDS.WaitForDbUp(DBSTARTUP_TIMEOUT_SECS)
	if err != nil {
		panic(err)
	}

	for {

		// grab the list of reports that have completes scanning but have not been sent out
		reports, err := vinylDS.GetAllCompletedUnsentReports(nil)
		if err != nil {
			log.Error(err, "Failed to get completed unsent reports")
		}

		if len(reports) > 0 {
			log.Debugf("Found %v completed, unsent reports..", len(reports))

			// index the reports by batch
			batchedReports := map[int64][]db.BatchedReport{}
			for _, rep := range reports {
				if _, ok := batchedReports[rep.BatchID]; !ok {
					batchedReports[rep.BatchID] = []db.BatchedReport{}
				}
				batchedReports[rep.BatchID] = append(batchedReports[rep.BatchID], rep)
			}

			// process a batch at a time and then mark the batch as reported.
			for batchID, reports := range batchedReports {
				sendFailed := false
				for _, report := range reports {
					log.Debugf("Processing report with ID: %v", report.ReportID)

					// get the details for the report (skus). if there are no skus attached then the report
					// has no valid results and can be deleted.
					skus, err := vinylDS.GetSkusForReport(nil, report.ReportID)
					if err != nil {
						log.Error(err, "failed to get skus for report %v", report.ReportID)
						sendFailed = true
						continue
					}
					log.Debugf("Report with ID: %v has %v skus attached", report.ReportID, len(skus))

					if len(skus) > 0 {
						// send the report to the watching user as an email listing the skus
						err = buildAndSendEmail(skus, report.UserEmail, report.UserName)
						if err != nil {
							log.Error(err, "Failed to send results email to %v", report.UserEmail)
							sendFailed = true
							continue
						}
						// mark the report as sent (we're done with it
						err = vinylDS.MarkReportSent(nil, report.ReportID)
						if err != nil {
							log.Error(err, "Failed to set report as sent")
						}
					} else {
						log.Debugf("Report %v for batch %v, user %v has no SKUs found. Deleting the report..", report.ReportID, batchID, report.UserEmail)
						err = vinylDS.DeleteReport(nil, report.ReportID)
						if err != nil {
							log.Error(err, "Failed to delete report %v  for user %s", report.ReportID, report.UserEmail)
						}
					}
				}
				// mark the whole batch as reported (we are done with it)
				if !sendFailed {
					err = vinylDS.MarkBatchReported(nil, batchID)
					if err != nil {
						log.Error(err, "Failed to mark batch as reported")
					}
				}
			}
		}
		time.Sleep(time.Duration(DELAY_BETWEEN_REPORTING_POLLS_SECS) * time.Second)
	}
}

func renderResultsRow(image string, artist string, titleUrl string, title string, price string, retailer string, retailerUrl string) (string, error) {
	htmlOut := "<tr>\n"
	htmlOut += fmt.Sprintf("<td><img width=\"150px\" height=\"150px\" src=\"%s\"/></td>\n", image)
	htmlOut += fmt.Sprintf("<td>%s<br>\n", artist)
	if titleUrl != "" {
		htmlOut += fmt.Sprintf("<a href=\"%s\">%s</a><br>\n", titleUrl, title)
	} else {
		htmlOut += fmt.Sprintf("%s<br>\n", title)
	}
	if retailerUrl != "" {
		htmlOut += fmt.Sprintf("<a href=\"%s\">%s</a><br>", retailerUrl, retailer)
	} else {
		htmlOut += fmt.Sprintf("%s<br>", retailer)
	}
	htmlOut += fmt.Sprintf("%s</td>\n", price)
	htmlOut += "</tr>\n"
	return htmlOut, nil
}

func buildAndSendEmail(skus []retailers.SKU, userEmail string, userName string) error {
	subject := "New vinyl releases found by new engine"
	message := ""

	artistSkus := map[string][]retailers.SKU{}
	retailerSkus := map[string][]retailers.SKU{}

	for _, sku := range skus {
		log.Debugf("Processing SKU %v for report to %v", sku.Name, userEmail)
		if _, ok := artistSkus[sku.Artist]; !ok {
			artistSkus[sku.Artist] = []retailers.SKU{}
		}
		if _, ok := retailerSkus[sku.Retailer]; !ok {
			retailerSkus[sku.Retailer] = []retailers.SKU{}
		}
		s := artistSkus[sku.Artist]
		s = append(s, sku)
		artistSkus[sku.Artist] = s
		s = retailerSkus[sku.Retailer]
		s = append(s, sku)
		retailerSkus[sku.Retailer] = s
	}

	for artist, skus := range artistSkus {
		message += "<h4>" + artist + "</h4>\n"
		message += "<table>\n"
		for _, sku := range skus {
			log.Debugf("Processing SKU %v for artist %v, report to %v", sku.Name, artist, userEmail)
			row, err := renderResultsRow(sku.Image, sku.Artist, sku.Url, sku.Name, sku.Price, sku.Retailer, sku.RetailerUrl)
			if err != nil {
				return errors.Wrapf(err, "Failed to construct report email message")
			}
			message += row
		}
		message += "</table>"
	}
	return email.SendEmail(userEmail, subject, message)
}
