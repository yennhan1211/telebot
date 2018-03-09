package main

import (
    "log"
    "net/http"
    "sync"
    "bytes"
    "os"
    "os/signal"
    "time"
    "encoding/json"
    "strings"
    "fmt"
    "strconv"
    "path"
    "flag"
    "bufio"
    "syscall"

    "github.com/go-telegram-bot-api/telegram-bot-api"
    "github.com/kyokomi/emoji"
)

const MARKETCAP_URL = "https://api.coinmarketcap.com/v1/global/?convert=usd"
const BINANCE_URL = "https://www.binance.com/api/v1/ticker/price"
const BITTREX_URL = "https://bittrex.com/api/v1.1/public/getmarketsummaries"

var mlog * log.Logger
var wait_update sync.WaitGroup
var wait_display sync.WaitGroup

type binancePair struct {
    Symbol      string      `json:"symbol"`
    Price       string      `json:"price"`
}

type mcapCoinObj struct {
    ID                      string      `json:"id"`
    Name                    string      `json:"name"`
    Symbol                  string      `json:"symbol"`
    Rank                    string      `json:"rank"`
    Price_usd               string      `json:"price_usd"`
    Price_btc               string      `json:"price_btc"`
    Volume_24_usd           string      `json:"24_volume_usd"`
    Market_cap_usd          string      `json:"market_cap_usd"`
    Available_supply        string      `json:"available_supply"`
    Total_supply            string      `json:"total_supply"`
    Max_supply              string      `json:"max_supply"`
    Percent_change_1h       string      `json:"percent_change_1h"`
    Percent_change_24h      string      `json:"percent_change_24h"`
    Percent_change_7d       string      `json:"percent_change_7d"`
    Last_updated            string      `json:"last_updated"`
}

// {
//     "total_market_cap_usd": 442551712698.0,
//     "total_24h_volume_usd": 16859016718.0,
//     "bitcoin_percentage_of_market_cap": 41.65,
//     "active_currencies": 914,
//     "active_assets": 627,
//     "active_markets": 9121,
//     "last_updated": 1520389466
// }
type totalMarketCap struct {
    Total_market_cap_usd                        float64      `json:"total_market_cap_usd"`
    Total_24h_volume_usd                        float64      `json:"total_24h_volume_usd"`
    Bitcoin_percentage_of_market_cap            float64      `json:"bitcoin_percentage_of_market_cap"`
    Active_currencies                           float64      `json:"active_currencies"`
    Active_markets                              float64      `json:"Active_markets"`
    Last_updated                                float64      `json:"last_updated"`
}

// {"MarketName":"BTC-1ST","High":0.00003209,"Low":0.00002803,"Volume":2184446.03744252,"Last":0.00003029,"BaseVolume":66.18579856,
// "TimeStamp":"2018-03-07T01:15:18.597","Bid":0.00003029,"Ask":0.00003040,"OpenBuyOrders":188,"OpenSellOrders":2255,
// "PrevDay":0.00002950,"Created":"2017-06-06T01:22:35.727"}
type bittrexMarket struct {
    MarketName          string          `json:"MarketName"`
    High                float64          `json:"High"`
    Low                 float64          `json:"Low"`
    Volume              float64          `json:"Volume"`
    Last                float64          `json:"Last"`
    BaseVolume          float64          `json:"BaseVolume"`
    TimeStamp           string          `json:"TimeStamp"`
    Bid                 float64          `json:"Bid"`
    Ask                 float64          `json:"Ask"`
    OpenBuyOrders       float64          `json:"OpenBuyOrders"`
    OpenSellOrders      float64          `json:"OpenSellOrders"`
    PrevDay             float64          `json:"PrevDay"`
    Created             string          `json:"PrevDay"`
}

type bittrexResp struct {
    Success             bool                `json:"success"`
    Message             string              `json:"message"`
    Result              []bittrexMarket     `json:"result"`
}

const PAIR_REPLY = `
:point_right:  *%s*
    :point_right: Binance:      %s
    :point_right: Bittrex:       %s
`

const TROLL_DO = `
:point_right: *ĐĨ*
:point_right: *ĐỘ*
:point_right: *ĐẠI*
:point_right: *ĐẦN*
:point_right: *ĐỘN*
`

const TOTAL_MCAP = `
*Total market cap:*     %.1f USD
*BTC percent:*             %.2f %%
*Total volume 24h:*     %.1f USD
`

const HELP_STR = `
:point_right: "/coinsymbol"        To get price on BTC except BTC. Eg: /xvg
                                    Specific pair are accepted. Eg: /bnbusdt
:point_right: "/mcap"               To get total market cap by USD
`

var results []binancePair
var total_mcap totalMarketCap
var mcap_results []mcapCoinObj
var bittrex_results bittrexResp

func removePidFile(filePath string) {
    mlog.Printf("Remove pid file %s", filePath)
    err := os.Remove(filePath)
    if err != nil {
        mlog.Printf("Cannot remove pid file %+v", err)
    }
}

func createPidFile(filePath string) bool {
    basepath := path.Dir(filePath)

    errf := os.MkdirAll(basepath, 0777)
    if errf != nil {
        mlog.Printf("Cannot create pid folder %+v", errf)
        return false
    }

    f, err := os.Create(filePath)

    if err != nil {
        mlog.Printf("Cannot create pid file %+v", err)
        return false
    }

    s := strconv.Itoa(os.Getpid())
    w := bufio.NewWriter(f)
    w.WriteString(s)
    w.Flush()
    f.Close()

    return true
}

func requestUrl(mUrl string) ([]byte, error) {
    var err error
    var req *http.Request

    var byteResp = make([]byte, 0, 0)

    req, err = http.NewRequest("GET", mUrl, nil)
    if err != nil {
        // c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot create new request %s", err.Error()))
        return byteResp, err
    }

    req.Header.Set("Accept", "application/json")
    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        // c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot send request %s", err.Error()))
        // mlog.Println(err)
        return byteResp, err
    }
    defer res.Body.Close()

    body := &bytes.Buffer{}
    _, err = body.ReadFrom(res.Body)
    if err != nil {
        // c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err.Error()))
        mlog.Println(err)
        return byteResp, err
    }

    return body.Bytes(), nil
}


func updateData() {
    // var err error
    // var req *http.Request

    // mlog.Println("Update data from Binance")
    // req, err = http.NewRequest("GET", "https://www.binance.com/api/v1/ticker/price", nil)
    // if err != nil {
    //     // c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot create new request %s", err.Error()))
    //     mlog.Println(err)
    //     return
    // }

    // req.Header.Set("Accept", "application/json")
    // client := &http.Client{}
    // res, err := client.Do(req)
    // if err != nil {
    //     // c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot send request %s", err.Error()))
    //     mlog.Println(err)
    //     return
    // }
    // defer res.Body.Close()

    // body := &bytes.Buffer{}
    // _, err = body.ReadFrom(res.Body)
    // if err != nil {
    //     // c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err.Error()))
    //     mlog.Println(err)
    //     return
    // }
    byte_result_binance, err := requestUrl("https://www.binance.com/api/v1/ticker/price")
    if err != nil {
        mlog.Println(err)
        return
    }

    // byte_result_mcap

    byte_result_mcap, err := requestUrl(MARKETCAP_URL)
    if err != nil {
        mlog.Println(err)
        return
    }

    byte_result_bittrex, err := requestUrl(BITTREX_URL)
    if err != nil {
        mlog.Println(err)
        return
    }

// cm.mutex.Lock()
// defer cm.mutex.Unlock()
    wait_update.Add(1)

    wait_display.Wait() // wait for all read task finish
    // results = bytes.NewBuffer()

    err = json.Unmarshal(byte_result_binance, &results)

    if err != nil {
        mlog.Println(string(byte_result_binance))
        mlog.Println(err)
    }

    err = json.Unmarshal(byte_result_mcap, &total_mcap)

    if err != nil {
        mlog.Println(string(byte_result_mcap))
        mlog.Println(err)
    }

    err = json.Unmarshal(byte_result_bittrex, &bittrex_results)

    if err != nil {
        mlog.Println(string(byte_result_bittrex))
        mlog.Println(err)
    }

    wait_update.Done()
}

func botHandler(bot * tgbotapi.BotAPI, update tgbotapi.Update) {
    // var foundResult = false

    wait_update.Wait() // wait for update task finish

    incomingMsg := ""

    if update.Message != nil {
        incomingMsg = strings.Replace(update.Message.Text, "/", "", -1)
    } else if update.EditedMessage != nil {
        incomingMsg = strings.Replace(update.EditedMessage.Text, "/", "", -1)
    } else if update.ChannelPost != nil {
        mlog.Println(update.ChannelPost.Text)
        return
    } else if update.EditedChannelPost != nil {
        mlog.Println(update.EditedChannelPost.Text)
        return
    }

    if incomingMsg == "" {
        mlog.Println("Empty Message")
        return
    }

    wait_display.Add(1)

    // foundPair := binancePair{}

    foundBnB := -1
    foundBittrex := -1

    incomingMsg = strings.ToUpper(incomingMsg)
    replyMsg:= ""
    switch(incomingMsg) {
        case "5D":
            replyMsg = emoji.Sprintf(TROLL_DO)
            break
        case "MCAP":
            replyMsg = fmt.Sprintf(TOTAL_MCAP, total_mcap.Total_market_cap_usd, total_mcap.Bitcoin_percentage_of_market_cap, total_mcap.Total_24h_volume_usd)
            break
        case "HELP":
            replyMsg = emoji.Sprintf(HELP_STR)
            break

        default:
            for i, bPair := range(results) {
                if bPair.Symbol == (incomingMsg + "BTC") || bPair.Symbol == (incomingMsg + "USDT") || bPair.Symbol == (incomingMsg){
                    // foundResult = true
                    // foundPair.Symbol = bPair.Symbol
                    // foundPair.Price = bPair.Price
                    foundBnB = i
                    break
                }
            }
            for i, btPair := range(bittrex_results.Result) {
                if btPair.MarketName == ("BTC-" + incomingMsg) {
                    // foundResult = true
                    // foundPair.Symbol = bPair.Symbol
                    // foundPair.Price = bPair.Price
                    foundBittrex = i
                    break
                }
            }
            // if foundResult {
            //     replyMsg = emoji.Sprintf(PAIR_REPLY, foundPair.Symbol, foundPair.Price, "Not listed")
            // } else {
            //     replyMsg = fmt.Sprintf("Command not found -> %s", incomingMsg)
            // }
            priceBnB := ""
            priceBittrex := ""
            if foundBnB != -1 {
                priceBnB = results[foundBnB].Price
            } else {
                priceBnB = "Not listed"
            }

            if foundBittrex != -1 {
                priceBittrex = strconv.FormatFloat(bittrex_results.Result[foundBittrex].Last, 'f', -1, 64)
            } else {
                priceBittrex = "Not listed"
            }
            replyMsg = emoji.Sprintf(PAIR_REPLY,incomingMsg, priceBnB, priceBittrex)
            break
    }
    //  || bPair.Symbol == strings.ToUpper(incomingMsg + "ETH")

    wait_display.Done()

    msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyMsg)
    msg.ParseMode = tgbotapi.ModeMarkdown
    msg.ReplyToMessageID = update.Message.MessageID

    bot.Send(msg)
    mlog.Printf("%+v", update.Message)

    // mlog.Println(replyMsg)
    // mlog.Println("Out goroutin")
}

func main() {
    pid_path := flag.String("p", "/var/run/telebot/telebot.pid", "bot PID file")
    log_path := flag.String("l", "", "log destination")


    mlog = log.New(os.Stderr, "INFO: ",log.Ldate|log.Ltime|log.Lshortfile)
    // mlog.SetOutput()

    if *log_path != "" {
        fLog ,err := os.OpenFile(*log_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            mlog.Printf("Cannot access log file %s\n", *log_path)
            // removePidFile(*pid_path)
            os.Exit(1)
        }
        mlog.SetOutput(fLog)
    }

    if ! createPidFile(*pid_path) {
        removePidFile(*pid_path)
        os.Exit(1)
    }

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

    bot, err := tgbotapi.NewBotAPI("541804902:AAF4O1fZcNZ8MHaRW-aBBXNuQicadtbCg4c")
    if err != nil {
        mlog.Println(err)
        removePidFile(*pid_path)
        os.Exit(1)
    }

    bot.Debug = false

    mlog.Printf("Authorized on account %s\n", bot.Self.UserName)

    cfg := tgbotapi.NewWebhookWithCert("https://code4food.net:8443/"+bot.Token, "YOURPUBLIC.pem")

    _, err = bot.SetWebhook(cfg)
    if err != nil {
        mlog.Println(err)
        removePidFile(*pid_path)
        os.Exit(1)
    }

    updates := bot.ListenForWebhook("/" + bot.Token)

    update_data_ticker := time.NewTicker(time.Duration(30)* time.Second)
    quit_update_data_ticker := make(chan struct{})
    // call fisrt time
    go updateData()
    // run checklic every 30s
    go func() {
        for {
           select {
            case <- update_data_ticker.C:
                // check lic
                updateData()
            case <- quit_update_data_ticker:
                update_data_ticker.Stop()
                return
            }
        }
     }()

    go http.ListenAndServeTLS("0.0.0.0:8443", "YOURPUBLIC.pem", "YOURPRIVATE.key", nil)

    go func() {
        for update := range updates {
            // log.Printf("INFO: %+v\n", update)
            // log.Printf("INFO: %s\n", update.Message.Text)

            botHandler(bot, update)
        }
    }()

    for {
        sig := <-sigs
        if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGKILL {
            mlog.Println("Exiting telebot")
            removePidFile(*pid_path)
            os.Exit(0)
        } else if sig == syscall.SIGHUP {

        }
    }
}
