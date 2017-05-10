/* Vuls - Vulnerability Scanner
Copyright (C) 2016  Future Architect, Inc. Japan.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	c "github.com/future-architect/vuls/config"
	"github.com/future-architect/vuls/cveapi"
	"github.com/future-architect/vuls/models"
	"github.com/future-architect/vuls/oval"
	"github.com/future-architect/vuls/report"
	"github.com/future-architect/vuls/util"
	"github.com/google/subcommands"
)

// ReportCmd is subcommand for reporting
type ReportCmd struct {
	lang       string
	debug      bool
	debugSQL   bool
	configPath string
	resultsDir string
	logDir     string
	refreshCve bool

	cvssScoreOver      float64
	ignoreUnscoredCves bool
	httpProxy          string

	cvedbtype string
	cvedbpath string
	cvedbURL  string

	ovaldbtype string
	ovaldbpath string

	toSlack     bool
	toEMail     bool
	toLocalFile bool
	toS3        bool
	toAzureBlob bool

	formatJSON        bool
	formatXML         bool
	formatOneEMail    bool
	formatOneLineText bool
	formatShortText   bool
	formatFullText    bool

	gzip bool

	awsProfile  string
	awsS3Bucket string
	awsRegion   string

	azureAccount   string
	azureKey       string
	azureContainer string

	pipe bool

	diff bool
}

// Name return subcommand name
func (*ReportCmd) Name() string { return "report" }

// Synopsis return synopsis
func (*ReportCmd) Synopsis() string { return "Reporting" }

// Usage return usage
func (*ReportCmd) Usage() string {
	return `report:
	report
		[-lang=en|ja]
		[-config=/path/to/config.toml]
		[-results-dir=/path/to/results]
		[-log-dir=/path/to/log]
		[-refresh-cve]
		[-cvedb-type=sqlite3|mysql]
		[-cvedb-path=/path/to/cve.sqlite3]
		[-cvedb-url=http://127.0.0.1:1323 or mysql connection string]
		[-cvss-over=7]
		[-diff]
		[-ignore-unscored-cves]
		[-to-email]
		[-to-slack]
		[-to-localfile]
		[-to-s3]
		[-to-azure-blob]
		[-format-json]
		[-format-xml]
		[-format-one-email]
		[-format-one-line-text]
		[-format-short-text]
		[-format-full-text]
		[-gzip]
		[-aws-profile=default]
		[-aws-region=us-west-2]
		[-aws-s3-bucket=bucket_name]
		[-azure-account=accout]
		[-azure-key=key]
		[-azure-container=container]
		[-http-proxy=http://192.168.0.1:8080]
		[-debug]
		[-debug-sql]
		[-pipe]

		[SERVER]...
`
}

// SetFlags set flag
func (p *ReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.lang, "lang", "en", "[en|ja]")
	f.BoolVar(&p.debug, "debug", false, "debug mode")
	f.BoolVar(&p.debugSQL, "debug-sql", false, "SQL debug mode")

	wd, _ := os.Getwd()

	defaultConfPath := filepath.Join(wd, "config.toml")
	f.StringVar(&p.configPath, "config", defaultConfPath, "/path/to/toml")

	defaultResultsDir := filepath.Join(wd, "results")
	f.StringVar(&p.resultsDir, "results-dir", defaultResultsDir, "/path/to/results")

	defaultLogDir := util.GetDefaultLogDir()
	f.StringVar(&p.logDir, "log-dir", defaultLogDir, "/path/to/log")

	f.BoolVar(
		&p.refreshCve,
		"refresh-cve",
		false,
		"Refresh CVE information in JSON file under results dir")

	f.StringVar(
		&p.cvedbtype,
		"cvedb-type",
		"sqlite3",
		"DB type for fetching CVE dictionary (sqlite3 or mysql)")

	defaultCveDBPath := filepath.Join(wd, "cve.sqlite3")
	f.StringVar(
		&p.cvedbpath,
		"cvedb-path",
		defaultCveDBPath,
		"/path/to/sqlite3 (For get cve detail from cve.sqlite3)")

	f.StringVar(
		&p.ovaldbtype,
		"ovaldb-type",
		"sqlite3",
		"DB type for fetching OVAL dictionary (sqlite3 or mysql)")

	defaultOvalDBPath := filepath.Join(wd, "oval.sqlite3")
	f.StringVar(
		&p.ovaldbpath,
		"ovaldb-path",
		defaultOvalDBPath,
		"/path/to/sqlite3 (For get oval detail from oval.sqlite3)")

	f.StringVar(
		&p.cvedbURL,
		"cvedb-url",
		"",
		"http://cve-dictionary.com:8080 or mysql connection string")

	f.Float64Var(
		&p.cvssScoreOver,
		"cvss-over",
		0,
		"-cvss-over=6.5 means reporting CVSS Score 6.5 and over (default: 0 (means report all))")

	f.BoolVar(&p.diff,
		"diff",
		false,
		fmt.Sprintf("Difference between previous result and current result "))

	f.BoolVar(
		&p.ignoreUnscoredCves,
		"ignore-unscored-cves",
		false,
		"Don't report the unscored CVEs")

	f.StringVar(
		&p.httpProxy,
		"http-proxy",
		"",
		"http://proxy-url:port (default: empty)")

	f.BoolVar(&p.formatJSON,
		"format-json",
		false,
		fmt.Sprintf("JSON format"))

	f.BoolVar(&p.formatXML,
		"format-xml",
		false,
		fmt.Sprintf("XML format"))

	f.BoolVar(&p.formatOneEMail,
		"format-one-email",
		false,
		"Send all the host report via only one EMail (Specify with -to-email)")

	f.BoolVar(&p.formatOneLineText,
		"format-one-line-text",
		false,
		fmt.Sprintf("One line summary in plain text"))

	f.BoolVar(&p.formatShortText,
		"format-short-text",
		false,
		fmt.Sprintf("Summary in plain text"))

	f.BoolVar(&p.formatFullText,
		"format-full-text",
		false,
		fmt.Sprintf("Detail report in plain text"))

	f.BoolVar(&p.gzip, "gzip", false, "gzip compression")

	f.BoolVar(&p.toSlack, "to-slack", false, "Send report via Slack")
	f.BoolVar(&p.toEMail, "to-email", false, "Send report via Email")
	f.BoolVar(&p.toLocalFile,
		"to-localfile",
		false,
		fmt.Sprintf("Write report to localfile"))

	f.BoolVar(&p.toS3,
		"to-s3",
		false,
		"Write report to S3 (bucket/yyyyMMdd_HHmm/servername.json/xml/txt)")
	f.StringVar(&p.awsProfile, "aws-profile", "default", "AWS profile to use")
	f.StringVar(&p.awsRegion, "aws-region", "us-east-1", "AWS region to use")
	f.StringVar(&p.awsS3Bucket, "aws-s3-bucket", "", "S3 bucket name")

	f.BoolVar(&p.toAzureBlob,
		"to-azure-blob",
		false,
		"Write report to Azure Storage blob (container/yyyyMMdd_HHmm/servername.json/xml/txt)")
	f.StringVar(&p.azureAccount,
		"azure-account",
		"",
		"Azure account name to use. AZURE_STORAGE_ACCOUNT environment variable is used if not specified")
	f.StringVar(&p.azureKey,
		"azure-key",
		"",
		"Azure account key to use. AZURE_STORAGE_ACCESS_KEY environment variable is used if not specified")
	f.StringVar(&p.azureContainer, "azure-container", "", "Azure storage container name")

	f.BoolVar(
		&p.pipe,
		"pipe",
		false,
		"Use args passed via PIPE")
}

// Execute execute
func (p *ReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.Conf.Debug = p.debug
	c.Conf.DebugSQL = p.debugSQL
	c.Conf.LogDir = p.logDir
	util.Log = util.NewCustomLogger(c.ServerInfo{})

	if err := c.Load(p.configPath, ""); err != nil {
		util.Log.Errorf("Error loading %s, %s", p.configPath, err)
		return subcommands.ExitUsageError
	}

	c.Conf.Lang = p.lang
	c.Conf.ResultsDir = p.resultsDir
	c.Conf.CveDBType = p.cvedbtype
	c.Conf.CveDBPath = p.cvedbpath
	c.Conf.CveDBURL = p.cvedbURL
	c.Conf.OvalDBType = p.ovaldbtype
	c.Conf.OvalDBPath = p.ovaldbpath
	c.Conf.CvssScoreOver = p.cvssScoreOver
	c.Conf.IgnoreUnscoredCves = p.ignoreUnscoredCves
	c.Conf.HTTPProxy = p.httpProxy

	c.Conf.Pipe = p.pipe

	c.Conf.FormatXML = p.formatXML
	c.Conf.FormatJSON = p.formatJSON
	c.Conf.FormatOneEMail = p.formatOneEMail
	c.Conf.FormatOneLineText = p.formatOneLineText
	c.Conf.FormatShortText = p.formatShortText
	c.Conf.FormatFullText = p.formatFullText

	c.Conf.GZIP = p.gzip
	c.Conf.Diff = p.diff

	var dir string
	var err error
	if p.diff {
		dir, err = jsonDir([]string{})
	} else {
		dir, err = jsonDir(f.Args())
	}
	if err != nil {
		util.Log.Errorf("Failed to read from JSON: %s", err)
		return subcommands.ExitFailure
	}

	// report
	reports := []report.ResultWriter{
		report.StdoutWriter{},
	}

	if p.toSlack {
		reports = append(reports, report.SlackWriter{})
	}

	if p.toEMail {
		reports = append(reports, report.EMailWriter{})
	}

	if p.toLocalFile {
		reports = append(reports, report.LocalFileWriter{
			CurrentDir: dir,
		})
	}

	if p.toS3 {
		c.Conf.AwsRegion = p.awsRegion
		c.Conf.AwsProfile = p.awsProfile
		c.Conf.S3Bucket = p.awsS3Bucket
		if err := report.CheckIfBucketExists(); err != nil {
			util.Log.Errorf("Check if there is a bucket beforehand: %s, err: %s", c.Conf.S3Bucket, err)
			return subcommands.ExitUsageError
		}
		reports = append(reports, report.S3Writer{})
	}

	if p.toAzureBlob {
		c.Conf.AzureAccount = p.azureAccount
		if len(c.Conf.AzureAccount) == 0 {
			c.Conf.AzureAccount = os.Getenv("AZURE_STORAGE_ACCOUNT")
		}

		c.Conf.AzureKey = p.azureKey
		if len(c.Conf.AzureKey) == 0 {
			c.Conf.AzureKey = os.Getenv("AZURE_STORAGE_ACCESS_KEY")
		}

		c.Conf.AzureContainer = p.azureContainer
		if len(c.Conf.AzureContainer) == 0 {
			util.Log.Error("Azure storage container name is requied with --azure-container option")
			return subcommands.ExitUsageError
		}
		if err := report.CheckIfAzureContainerExists(); err != nil {
			util.Log.Errorf("Check if there is a container beforehand: %s, err: %s", c.Conf.AzureContainer, err)
			return subcommands.ExitUsageError
		}
		reports = append(reports, report.AzureBlobWriter{})
	}

	if !(p.formatJSON || p.formatOneLineText ||
		p.formatShortText || p.formatFullText || p.formatXML) {
		c.Conf.FormatShortText = true
	}

	util.Log.Info("Validating config...")
	if !c.Conf.ValidateOnReport() {
		return subcommands.ExitUsageError
	}
	if ok, err := cveapi.CveClient.CheckHealth(); !ok {
		util.Log.Errorf("CVE HTTP server is not running. err: %s", err)
		util.Log.Errorf("Run go-cve-dictionary as server mode before reporting or run with --cvedb-path option")
		return subcommands.ExitFailure
	}
	if c.Conf.CveDBURL != "" {
		util.Log.Infof("cve-dictionary: %s", c.Conf.CveDBURL)
	} else {
		if c.Conf.CveDBType == "sqlite3" {
			util.Log.Infof("cve-dictionary: %s", c.Conf.CveDBPath)
		}
	}

	rs, err := loadScanResults(dir)
	if err != nil {
		util.Log.Error(err)
		return subcommands.ExitFailure
	}
	util.Log.Infof("Loaded: %s", dir)

	var results []models.ScanResult
	for _, r := range rs {
		if p.refreshCve || needToRefreshCve(r) {
			util.Log.Debugf("need to refresh")
			if c.Conf.CveDBType == "sqlite3" && c.Conf.CveDBURL == "" {
				if _, err := os.Stat(c.Conf.CveDBPath); os.IsNotExist(err) {
					util.Log.Errorf("SQLite3 DB(CVE-Dictionary) is not exist: %s",
						c.Conf.CveDBPath)
					return subcommands.ExitFailure
				}
			}

			if err := fillCveInfoFromOvalDB(&r); err != nil {
				util.Log.Errorf("Failed to fill OVAL information: %s", err)
				return subcommands.ExitFailure
			}

			if err := fillCveInfoFromCveDB(&r); err != nil {
				util.Log.Errorf("Failed to fill CVE information: %s", err)
				return subcommands.ExitFailure
			}

			r.Lang = c.Conf.Lang
			if err := overwriteJSONFile(dir, r); err != nil {
				util.Log.Errorf("Failed to write JSON: %s", err)
				return subcommands.ExitFailure
			}
			results = append(results, r)
		} else {
			util.Log.Debugf("no need to refresh")
			results = append(results, r)
		}
	}

	if p.diff {
		previous, err := loadPrevious(results)
		if err != nil {
			util.Log.Error(err)
			return subcommands.ExitFailure
		}

		diff, err := diff(results, previous)
		if err != nil {
			util.Log.Error(err)
			return subcommands.ExitFailure
		}
		results = []models.ScanResult{}
		for _, r := range diff {
			if err := fillCveDetail(&r); err != nil {
				util.Log.Error(err)
				return subcommands.ExitFailure
			}
			results = append(results, r)
		}
	}

	var res models.ScanResults
	for _, r := range results {
		res = append(res, r.FilterByCvssOver())

		// TODO Add sort function to ScanResults

		//remove
		//  for _, vuln := range r.ScannedCves {
		//      //  if _, ok := vuln.CveContents.Get(models.NewCveContentType(r.Family)); !ok {
		//      //      pp.Printf("not in oval: %s %f\n%v\n",
		//      //          vuln.CveID, vuln.CveContents.CvssV2Score(), vuln.Packages)
		//      //  } else {
		//      //      fmt.Printf("    in oval: %s %f\n",
		//      //          vuln.CveID, vuln.CveContents.CvssV2Score())
		//      //  }
		//      //  if vuln.CveContents.CvssV2Score() < 0.1 &&
		//      //      vuln.CveContents.CvssV3Score() < 0.1 {
		//      //      pp.Println(vuln)
		//      //  }
		//  }
	}

	for _, w := range reports {
		if err := w.Write(res...); err != nil {
			util.Log.Errorf("Failed to report: %s", err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}

// fillCveDetail fetches NVD, JVN from CVE Database, and then set to fields.
func fillCveDetail(r *models.ScanResult) error {
	var cveIDs []string
	for _, v := range r.ScannedCves {
		cveIDs = append(cveIDs, v.CveID)
	}

	ds, err := cveapi.CveClient.FetchCveDetails(cveIDs)
	if err != nil {
		return err
	}
	for _, d := range ds {
		nvd := r.ConvertNvdToModel(d.CveID, d.Nvd)
		jvn := r.ConvertJvnToModel(d.CveID, d.Jvn)
		for cveID, vinfo := range r.ScannedCves {
			if vinfo.CveID == d.CveID {
				if vinfo.CveContents == nil {
					vinfo.CveContents = models.CveContents{}
				}
				for _, con := range []models.CveContent{*nvd, *jvn} {
					if !con.Empty() {
						vinfo.CveContents[con.Type] = con
					}
				}
				r.ScannedCves[cveID] = vinfo
				break
			}
		}
	}
	//TODO Remove
	//  sort.Slice(r.ScannedCves, func(i, j int) bool {
	//      if r.ScannedCves[j].CveContents.CvssV2Score() == r.ScannedCves[i].CveContents.CvssV2Score() {
	//          return r.ScannedCves[j].CveContents.CvssV2Score() < r.ScannedCves[i].CveContents.CvssV2Score()
	//      }
	//      return r.ScannedCves[j].CveContents.CvssV2Score() < r.ScannedCves[i].CveContents.CvssV2Score()
	//  })
	return nil
}

func fillCveInfoFromCveDB(r *models.ScanResult) error {
	sInfo := c.Conf.Servers[r.ServerName]
	if err := fillVulnByCpeNames(sInfo.CpeNames, r.ScannedCves); err != nil {
		return err
	}
	if err := fillCveDetail(r); err != nil {
		return err
	}
	return nil
}

func fillCveInfoFromOvalDB(r *models.ScanResult) error {
	var ovalClient oval.Client
	switch r.Family {
	case "ubuntu", "debian":
		ovalClient = oval.NewDebian()
	case "rhel", "centos":
		ovalClient = oval.NewRedhat()
	case "amazon", "oraclelinux", "Raspbian", "FreeBSD":
		//TODO implement OracleLinux
		return nil
	default:
		return fmt.Errorf("Oval %s is not implemented yet", r.Family)
	}
	if err := ovalClient.FillCveInfoFromOvalDB(r); err != nil {
		return err
	}
	return nil
}
