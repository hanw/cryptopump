package types

import (
	"database/sql"
	"time"

	"github.com/adshao/go-binance/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sdcoffey/techan"
	log "github.com/sirupsen/logrus"
)

// Order struct define an exchange order
type Order struct {
	ClientOrderID           string  `json:"clientOrderId"`
	CumulativeQuoteQuantity float64 `json:"cumulativeQuoteQty"`
	ExecutedQuantity        float64 `json:"executedQty"`
	OrderID                 int     `json:"orderId"`
	Price                   float64 `json:"price"`
	Side                    string  `json:"side"`
	Status                  string  `json:"status"`
	Symbol                  string  `json:"symbol"`
	TransactTime            int64   `json:"transactTime"`
	ThreadID                int
	ThreadIDSession         int
	OrderIDSource           int /* Used for logging purposes to define source OrderID for a sale */
}

// Kline struct define a kline
type Kline struct {
	OpenTime int64  `json:"openTime"`
	Open     string `json:"open"`
	High     string `json:"high"`
	Low      string `json:"low"`
	Close    string `json:"close"`
	Volume   string `json:"volume"`
}

// WsKline struct define websocket kline
type WsKline struct {
	StartTime            int64  `json:"t"`
	EndTime              int64  `json:"T"` /* Currently not in use */
	Symbol               string `json:"s"` /* Currently not in use */
	Interval             string `json:"i"` /* Currently not in use */
	FirstTradeID         int64  `json:"f"` /* Currently not in use */
	LastTradeID          int64  `json:"L"` /* Currently not in use */
	Open                 string `json:"o"`
	Close                string `json:"c"`
	High                 string `json:"h"`
	Low                  string `json:"l"`
	Volume               string `json:"v"`
	TradeNum             int64  `json:"n"` /* Currently not in use */
	IsFinal              bool   `json:"x"`
	QuoteVolume          string `json:"q"` /* Currently not in use */
	ActiveBuyVolume      string `json:"V"` /* Currently not in use */
	ActiveBuyQuoteVolume string `json:"Q"` /* Currently not in use */
}

// PriceChangeStats define price change stats
type PriceChangeStats struct {
	HighPrice string `json:"highPrice"`
	LowPrice  string `json:"lowPrice"`
}

// ExchangeInfo define exchange order size
type ExchangeInfo struct {
	MaxQuantity string `json:"maxQty"`
	MinQuantity string `json:"minQty"`
	StepSize    string `json:"stepSize"`
}

// Session struct define session elements
type Session struct {
	ThreadID             string /* Unique session ID for the thread */
	ThreadIDSession      string
	ThreadCount          int
	SellTransactionCount float64 /* Number of SELL transactions in the last 60 minutes */
	Symbol               string
	SymbolFiat           string
	SymbolFiatFunds      float64
	LastBuyTransactTime  time.Time /* This session variable stores the time of the last buy */
	LastSellCanceledTime time.Time /* This session variable stores the time of the cancelled sell */
	ConfigTemplate       int
	ForceBuy             bool             /* This boolean when True force BUY transaction */
	ForceSell            bool             /* This boolean when True force SELL transaction */
	ListenKey            string           /* Listen key for user stream service */
	MasterNode           bool             /* This boolean is true when Master Node is elected */
	TgBotAPI             *tgbotapi.BotAPI /* This variable holds Telegram session bot */
	Db                   *sql.DB          /* mySQL database connection */
	Clients              Client           /* Binance client connection */
	KlineData            []KlineData      /* kline data format for go-echart plotter */
	StopWs               bool             /* Control when to stop Ws Channels */
	Busy                 bool             /* Control wether buy/selling to allow graceful session exit */
	MinQuantity          float64          /* Defines the minimum quantity allowed by exchange */
	MaxQuantity          float64          /* Defines the maximum quantity allowed by exchange */
	StepSize             float64          /* Defines the intervals that a quantity can be increased/decreased by exchange */
}

// Client struct for client libraries
type Client struct {
	Binance *binance.Client
}

// WsHandler struct for websocket handlers for exchanges
type WsHandler struct {
	BinanceWsKline         func(event *binance.WsKlineEvent)      /* WsKlineServe serve websocket kline handler */
	BinanceWsBookTicker    func(event *binance.WsBookTickerEvent) /* WsBookTicker serve websocket kline handler */
	BinanceWsUserDataServe func(message []byte)                   /* WsUserDataServe serve user data handler with listen key */
}

// KlineData struct define kline retention for e-charts plotting
type KlineData struct {
	Date    int64
	Data    [4]float64
	Volumes float64
}

// Market struct define realtime market data
type Market struct {
	Rsi3                      float64            /* Relative Strength Index for 3 periods */
	Rsi7                      float64            /* Relative Strength Index for 7 periods */
	Rsi14                     float64            /* Relative Strength Index for 14 periods */
	MACD                      float64            /* Moving average convergence divergence */
	Price                     float64            /* Market Price */
	PriceChangeStatsHighPrice float64            /* High price for 1 period */
	PriceChangeStatsLowPrice  float64            /* Low price for 1 period */
	Direction                 int                /* Market Direction */
	TimeStamp                 time.Time          /* Time of last retrieved market Data */
	Series                    *techan.TimeSeries /* kline data format for technical analysis */
}

// Config struct for configuration
type Config struct {
	ThreadID                               string /* For index.html population */
	Apikey                                 string /* Exchange API Key */
	Secretkey                              string /* Exchange Secret Key */
	ApikeyTestNet                          string /* API key for exchange test network, used with launch.json */
	SecretkeyTestNet                       string /* Secret key for exchange test network, used with launch.json */
	Buy24hsHighpriceEntry                  float64
	BuyDirectionDown                       int
	BuyDirectionUp                         int
	BuyQuantityFiatUp                      float64
	BuyQuantityFiatDown                    float64
	BuyQuantityFiatInit                    float64
	BuyRepeatThresholdDown                 float64
	BuyRepeatThresholdDownSecond           float64
	BuyRepeatThresholdDownSecondStartCount int
	BuyRepeatThresholdUp                   float64
	BuyRsi7Entry                           float64
	BuyWait                                int /* Wait time between BUY transactions in seconds */
	ExchangeComission                      float64
	ProfitMin                              float64
	SellWaitBeforeCancel                   int     /* Wait time before cancelling a sale in seconds */
	SellWaitAfterCancel                    int     /* Wait time before selling after a cancel in seconds */
	SellToCover                            bool    /* Define if will sell to cover low funds */
	SellHoldOnRSI3                         float64 /* Hold sale if RSI3 above defined threshold */
	SymbolFiat                             string
	SymbolFiatStash                        float64
	Symbol                                 string
	TimeEnforce                            bool
	TimeStart                              string
	TimeStop                               string
	Debug                                  bool
	Exit                                   bool
	DryRun                                 bool        /* Dry Run mode */
	NewSession                             bool        /* Force a new session instead of resume */
	ConfigTemplateList                     interface{} /* List of configuration templates available in ./config folder */
	ExchangeName                           string      /* Exchange name */
	TestNet                                bool        /* Use Exchange TestNet */
	TgBotApikey                            string      /* Telegram bot API key */
	HTMLSnippet                            interface{} /* Store kline plotter graph for html output */
}

// OutboundAccountPosition Struct for User Data Streams for Binance
type OutboundAccountPosition struct {
	EventType  string     `json:"e"` /* Event type */
	EventTime  int64      `json:"E"` /* Event Time */
	LastUpdate int64      `json:"u"` /* Time of last account update */
	Balances   []Balances `json:"B"` /* Balances Array */
}

// Balances Struct for User Data Streams for Binance
type Balances struct {
	Asset  string `json:"a"` /* Asset */
	Free   string `json:"f"` /* Free */
	Locked string `json:"l"` /* Locked */
}

// ExecutionReport struct define exchange websocket transactions
type ExecutionReport struct {
	EventType             string `json:"e"` //Event type
	EventTime             int64  `json:"E"` //Event Time
	Symbol                string `json:"s"` //Symbol
	ClientOrderID         string `json:"c"` //Client order ID
	Side                  string `json:"S"` //Side
	OrderType             string `json:"o"` //Order type
	TimeInForce           string `json:"f"` //Time in force
	Quantity              string `json:"q"` //Order quantity
	Price                 string `json:"p"` //Order price
	StopPrice             string `json:"P"` //Stop price
	IcebergQuantity       string `json:"F"` //Iceberg quantity
	OrderListID           int64  `json:"g"` //OrderListId
	OriginalClientOrderID string `json:"C"` //Original client order ID; This is the ID of the order being canceled
	ExecutionType         string `json:"x"` //Current execution type
	Status                string `json:"X"` //Current order status
	OrderRejectReason     string `json:"r"` //Order reject reason; will be an error code.
	OrderID               int    `json:"i"` //Order ID
	LastExecutedQuantity  string `json:"l"` //Last executed quantity
	CumulativeQty         string `json:"z"` //Cumulative filled quantity
	LastExecutedPrice     string `json:"L"` //Last executed price
	ComissionAmount       string `json:"n"` //Commission amount
	ComissionAsset        string `json:"N"` //Commission asset
	TransactTime          int64  `json:"T"` //Transaction time
	TradeID               int    `json:"t"` //Trade ID
	Ignore0               int    `json:"I"` //Ignore
	IsOrderOnTheBook      bool   `json:"w"` //Is the order on the book?
	IsTradeMakerSide      bool   `json:"m"` //Is this trade the maker side?
	Ignore1               bool   `json:"M"` //Ignore
	OrderCreationTime     int64  `json:"O"` //Order creation time
	CumulativeQuoteQty    string `json:"Z"` //Cumulative quote asset transacted quantity
	LastQuoteQty          string `json:"Y"` //Last quote asset transacted quantity (i.e. lastPrice * lastQty)
	QuoteOrderQty         string `json:"Q"` //Quote Order Qty
}

// LogEntry struct
type LogEntry struct {
	Config   *Config   /* Config struct */
	Market   *Market   /* Market struct */
	Session  *Session  /* Session struct */
	Order    *Order    /* Order struct */
	Message  string    /* Error message */
	LogLevel log.Level /* Logrus log level */
}
