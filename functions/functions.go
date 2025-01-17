package functions

import (
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"os"
	"strconv"

	"cryptopump/types"

	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// StrToFloat64 function
/* This public function convert string to float64 */
func StrToFloat64(value string) (r float64) {

	var err error

	if r, err = strconv.ParseFloat(value, 8); err != nil {

		log.Fatal(err)

	}

	return r
}

// Float64ToStr function
/* This public function convert float64 to string with variable precision */
func Float64ToStr(value float64, prec int) string {

	return strconv.FormatFloat(value, 'f', prec, 64)

}

// IntToFloat64 convert Int to Float64
func IntToFloat64(value int) float64 {

	return float64(value)

}

// StrToInt convert string to int
func StrToInt(value string) (r int) {

	var err error

	if r, err = strconv.Atoi(value); err != nil {

		fmt.Println(err)

	}

	return r

}

// Logger is responsible for all system logging
func Logger(LogEntry *types.LogEntry) {

	var err error
	var filename string
	var file *os.File

	/* Log as JSON instead of the default ASCII formatter */
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableSorting:  false,
	})

	log.SetLevel(LogEntry.LogLevel) /* Define the log level for the entry */

	switch {
	case LogEntry.LogLevel == log.InfoLevel:

		filename = "cryptopump.log"

	case LogEntry.LogLevel == log.DebugLevel:

		filename = "cryptopump_debug.log"

	}

	/* io.Writer output set for file */
	if file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666); err != nil {

		log.Fatal(err)

	}

	log.SetOutput(file)

	switch {
	case LogEntry.LogLevel == log.InfoLevel:

		switch LogEntry.Message {
		case "UP", "DOWN", "INIT":

			log.WithFields(log.Fields{
				"threadID":  LogEntry.Session.ThreadID,
				"rsi3":      fmt.Sprintf("%.2f", LogEntry.Market.Rsi3),
				"rsi7":      fmt.Sprintf("%.2f", LogEntry.Market.Rsi7),
				"rsi14":     fmt.Sprintf("%.2f", LogEntry.Market.Rsi14),
				"MACD":      fmt.Sprintf("%.2f", LogEntry.Market.MACD),
				"high":      LogEntry.Market.PriceChangeStatsHighPrice,
				"direction": LogEntry.Market.Direction,
			}).Info(LogEntry.Message)

		case "BUY":

			log.WithFields(log.Fields{
				"threadID":   LogEntry.Session.ThreadID,
				"orderID":    LogEntry.Order.OrderID,
				"orderPrice": fmt.Sprintf("%.4f", LogEntry.Order.Price),
			}).Info(LogEntry.Message)

		case "SELL":

			log.WithFields(log.Fields{
				"threadID":      LogEntry.Session.ThreadID,
				"OrderIDSource": LogEntry.Order.OrderIDSource,
				"orderID":       LogEntry.Order.OrderID,
				"orderPrice":    fmt.Sprintf("%.4f", LogEntry.Order.Price),
			}).Info(LogEntry.Message)

		case "CANCELED":

			if LogEntry.Config.Debug {

				log.WithFields(log.Fields{
					"threadID":      LogEntry.Session.ThreadID,
					"OrderIDSource": LogEntry.Order.OrderIDSource,
					"orderID":       LogEntry.Order.OrderID,
				}).Info(LogEntry.Message)

			}

		default:

			log.WithFields(log.Fields{
				"threadID": LogEntry.Session.ThreadID,
			}).Info(LogEntry.Message)

		}

	case LogEntry.LogLevel == log.DebugLevel:

		log.WithFields(log.Fields{
			"threadID": LogEntry.Session.ThreadID,
			"orderID":  LogEntry.Order.OrderID,
		}).Debug(LogEntry.Message)

	}

}

// MustGetenv is a helper function for getting environment variables.
// Displays a warning if the environment variable is not set.
func MustGetenv(k string) string {

	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Warning: %s environment variable not set.\n", k)
	}

	return strings.ToLower(v)

}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {

	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}

// GetThreadID Return random thread ID
func GetThreadID() string {

	return xid.New().String()

}

/* Convert Strign to Time */
func stringToTime(str string) (r time.Time) {

	var err error

	if r, err = time.Parse(time.Kitchen, str); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

	}

	return r

}

// IsInTimeRange Check if time is in a specific range
func IsInTimeRange(startTimeString string, endTimeString string) bool {

	t := time.Now()
	timeNowString := t.Format(time.Kitchen)
	timeNow := stringToTime(timeNowString)
	start := stringToTime(startTimeString)
	end := stringToTime(endTimeString)

	if timeNow.Before(start) {

		return false

	}

	if timeNow.After(end) {

		return false

	}

	return true

}

// IsFundsAvailable Validate available funds to buy
func IsFundsAvailable(
	configData *types.Config,
	sessionData *types.Session) bool {

	return (sessionData.SymbolFiatFunds - configData.SymbolFiatStash) >= configData.BuyQuantityFiatDown

}

/* Select the correct html template based on sessionData */
func selectTemplate(
	sessionData *types.Session) (template string) {

	if sessionData.ThreadID == "" {

		template = "index.html"

	} else {

		template = "index_nostart.html"

	}

	return template

}

// ExecuteTemplate is responsible for executing any templates
func ExecuteTemplate(
	wr io.Writer,
	data interface{},
	sessionData *types.Session) {

	var tlp *template.Template
	var err error

	if tlp, err = template.ParseGlob("./templates/*"); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

		os.Exit(1)

	}

	if err = tlp.ExecuteTemplate(wr, selectTemplate(sessionData), data); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

		os.Exit(1)

	}

}

// GetFunctionName Retrieve current function name
func GetFunctionName() string {

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	return frame.Function

}

// LockThreadID Create lock for threadID
func LockThreadID(threadID string) bool {

	filename := threadID + ".lock"

	if _, err := os.Stat(filename); err == nil {

		return false

	} else if os.IsNotExist(err) {

		var file, err = os.Create(filename)
		if err != nil {
			return false
		}

		file.Close()

		return true

	}

	return false

}

// GetPort Determine port for HTTP service.
func GetPort() (port string) {

	port = os.Getenv("PORT")

	if port == "" {

		port = "8080"

	}

	for {

		if l, err := net.Listen("tcp", ":"+port); err != nil {

			port = Float64ToStr((StrToFloat64(port) + 1), 0)

		} else {

			l.Close()
			break

		}

	}

	return port

}

// DeleteConfigFile Delete configuration file for ThreadID
func DeleteConfigFile(sessionData *types.Session) {

	filename := sessionData.ThreadID + ".yml"
	path := "./config/"

	if err := os.Remove(path + filename); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

		return

	}

}

// GetConfigData Retrieve or create config file based on ThreadID
func GetConfigData(
	sessionData *types.Session) *types.Config {

	configData := loadConfigData(sessionData)

	if sessionData.ThreadID != "" {

		filename := sessionData.ThreadID + ".yml"
		writePath := "./config/"

		if _, err := os.Stat(writePath + filename); err == nil {

			/* Test for existing ThreadID config file and load configuration */
			viper.SetConfigFile(writePath + filename)

			if err := viper.ReadInConfig(); err != nil {

				Logger(&types.LogEntry{
					Config:   nil,
					Market:   nil,
					Session:  nil,
					Order:    &types.Order{},
					Message:  GetFunctionName() + " - " + err.Error(),
					LogLevel: log.DebugLevel,
				})

			}

			configData = loadConfigData(sessionData)

		} else if os.IsNotExist(err) {

			/* Create new ThreadID config file and load configuration */
			if err := viper.WriteConfigAs(writePath + filename); err != nil {

				Logger(&types.LogEntry{
					Config:   nil,
					Market:   nil,
					Session:  nil,
					Order:    &types.Order{},
					Message:  GetFunctionName() + " - " + err.Error(),
					LogLevel: log.DebugLevel,
				})

			}

			viper.SetConfigFile(writePath + filename)

			if err := viper.ReadInConfig(); err != nil {

				Logger(&types.LogEntry{
					Config:   nil,
					Market:   nil,
					Session:  nil,
					Order:    &types.Order{},
					Message:  GetFunctionName() + " - " + err.Error(),
					LogLevel: log.DebugLevel,
				})

			}

			configData = loadConfigData(sessionData)

		}

	}

	return configData

}

/* This function retrieve the list of configuration files under the root config folder.
.yaml files are considered configuration files. */
func getConfigTemplateList(sessionData *types.Session) []string {

	var files []string
	files = append(files, "-")

	root := "./config"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".yml" {
			files = append(files, info.Name())
		}

		return nil
	})

	if err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

		os.Exit(1)

	}

	return files

}

// LoadConfigTemplate Load the selected configuration template
// Three's a BUG where it only works before the first UPDATE
func LoadConfigTemplate(
	sessionData *types.Session) *types.Config {

	var filename string

	/* Retrieve the list of configuration templates */
	files := getConfigTemplateList(sessionData)

	/* Iterate configuration templates to match the selection */
	for key, file := range files {
		if key == sessionData.ConfigTemplate {
			filename = file
		}
	}

	filenameOld := viper.ConfigFileUsed()

	/* Set selected template as current config and load settings and return configData*/
	viper.SetConfigFile("./config/" + filename)
	if err := viper.ReadInConfig(); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

	}

	configData := loadConfigData(sessionData)

	/* Set origina template as current config */
	viper.SetConfigFile(filenameOld)
	if err := viper.ReadInConfig(); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

	}

	return configData

}

/* This routine load viper configuration data into map[string]interface{} */
func loadConfigData(
	sessionData *types.Session) *types.Config {

	configData := &types.Config{
		ThreadID:                               sessionData.ThreadID, /* For index.html population */
		Apikey:                                 viper.GetString("config.apiKey"),
		Secretkey:                              viper.GetString("config.secretKey"),
		ApikeyTestNet:                          viper.GetString("config.apiKeyTestNet"),    /* API key for exchange test network, used with launch.json */
		SecretkeyTestNet:                       viper.GetString("config.secretKeyTestNet"), /* Secret key for exchange test network, used with launch.json */
		Buy24hsHighpriceEntry:                  viper.GetFloat64("config.buy_24hs_highprice_entry"),
		BuyDirectionDown:                       viper.GetInt("config.buy_direction_down"),
		BuyDirectionUp:                         viper.GetInt("config.buy_direction_up"),
		BuyQuantityFiatUp:                      viper.GetFloat64("config.buy_quantity_fiat_up"),
		BuyQuantityFiatDown:                    viper.GetFloat64("config.buy_quantity_fiat_down"),
		BuyQuantityFiatInit:                    viper.GetFloat64("config.buy_quantity_fiat_init"),
		BuyRepeatThresholdDown:                 viper.GetFloat64("config.buy_repeat_threshold_down"),
		BuyRepeatThresholdDownSecond:           viper.GetFloat64("config.buy_repeat_threshold_down_second"),
		BuyRepeatThresholdDownSecondStartCount: viper.GetInt("config.buy_repeat_threshold_down_second_start_count"),
		BuyRepeatThresholdUp:                   viper.GetFloat64("config.buy_repeat_threshold_up"),
		BuyRsi7Entry:                           viper.GetFloat64("config.buy_rsi7_entry"),
		BuyWait:                                viper.GetInt("config.buy_wait"),
		ExchangeComission:                      viper.GetFloat64("config.exchange_comission"),
		ExchangeName:                           viper.GetString("config.exchangename"),
		ProfitMin:                              viper.GetFloat64("config.profit_min"),
		SellWaitBeforeCancel:                   viper.GetInt("config.sellwaitbeforecancel"),
		SellWaitAfterCancel:                    viper.GetInt("config.sellwaitaftercancel"),
		SellToCover:                            viper.GetBool("config.selltocover"),
		SellHoldOnRSI3:                         viper.GetFloat64("config.sellholdonrsi3"),
		SymbolFiat:                             viper.GetString("config.symbol_fiat"),
		SymbolFiatStash:                        viper.GetFloat64("config.symbol_fiat_stash"),
		Symbol:                                 viper.GetString("config.symbol"),
		TimeEnforce:                            viper.GetBool("config.time_enforce"),
		TimeStart:                              viper.GetString("config.time_start"),
		TimeStop:                               viper.GetString("config.time_stop"),
		TestNet:                                viper.GetBool("config.testnet"),
		TgBotApikey:                            viper.GetString("config.tgbotapikey"),
		Debug:                                  viper.GetBool("config.debug"),
		Exit:                                   viper.GetBool("config.exit"),
		DryRun:                                 viper.GetBool("config.dryrun"),
		NewSession:                             viper.GetBool("config.newsession"),
		ConfigTemplateList:                     getConfigTemplateList(sessionData),
	}

	return configData

}

// SaveConfigData save viper configuration from html
func SaveConfigData(
	r *http.Request,
	sessionData *types.Session) {

	viper.Set("config.buy_24hs_highprice_entry", r.PostFormValue("buy24hsHighpriceEntry"))
	viper.Set("config.buy_direction_down", r.PostFormValue("buyDirectionDown"))
	viper.Set("config.buy_direction_up", r.PostFormValue("buyDirectionUp"))
	viper.Set("config.buy_quantity_fiat_up", r.PostFormValue("buyQuantityFiatUp"))
	viper.Set("config.buy_quantity_fiat_down", r.PostFormValue("buyQuantityFiatDown"))
	viper.Set("config.buy_quantity_fiat_init", r.PostFormValue("buyQuantityFiatInit"))
	viper.Set("config.buy_rsi7_entry", r.PostFormValue("buyRsi7Entry"))
	viper.Set("config.buy_wait", r.PostFormValue("buyWait"))
	viper.Set("config.buy_repeat_threshold_down", r.PostFormValue("buyRepeatThresholdDown"))
	viper.Set("config.buy_repeat_threshold_down_second", r.PostFormValue("buyRepeatThresholdDownSecond"))
	viper.Set("config.buy_repeat_threshold_down_second_start_count", r.PostFormValue("buyRepeatThresholdDownSecondStartCount"))
	viper.Set("config.buy_repeat_threshold_up", r.PostFormValue("buyRepeatThresholdUp"))
	viper.Set("config.exchange_comission", r.PostFormValue("exchangeComission"))
	viper.Set("config.exchangename", r.PostFormValue("exchangename"))
	viper.Set("config.profit_min", r.PostFormValue("profitMin"))
	viper.Set("config.sellwaitbeforecancel", r.PostFormValue("sellwaitbeforecancel"))
	viper.Set("config.sellwaitaftercancel", r.PostFormValue("sellwaitaftercancel"))
	viper.Set("config.selltocover", r.PostFormValue("selltocover"))
	viper.Set("config.sellholdonrsi3", r.PostFormValue("sellholdonrsi3"))
	viper.Set("config.symbol", r.PostFormValue("symbol"))
	viper.Set("config.symbol_fiat", r.PostFormValue("symbol_fiat"))
	viper.Set("config.symbol_fiat_stash", r.PostFormValue("symbolFiatStash"))
	viper.Set("config.time_enforce", r.PostFormValue("timeEnforce"))
	viper.Set("config.time_start", r.PostFormValue("timeStart"))
	viper.Set("config.time_stop", r.PostFormValue("timeStop"))
	viper.Set("config.testnet", r.PostFormValue("testnet"))
	viper.Set("config.debug", r.PostFormValue("debug"))
	viper.Set("config.exit", r.PostFormValue("exit"))
	viper.Set("config.dryrun", r.PostFormValue("dryrun"))
	viper.Set("config.newsession", r.PostFormValue(("newsession")))

	if err := viper.WriteConfig(); err != nil {

		Logger(&types.LogEntry{
			Config:   nil,
			Market:   nil,
			Session:  nil,
			Order:    &types.Order{},
			Message:  GetFunctionName() + " - " + err.Error(),
			LogLevel: log.DebugLevel,
		})

		os.Exit(1)
	}

}
