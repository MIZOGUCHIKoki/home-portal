package main

import (
    "fmt"
    "log"
    "os"
)

func main() {
    mode := "app"
    if len(os.Args) > 1 {
        mode = os.Args[1]
    }

    db := ConnectDB()
    defer db.Close()

    switch mode {
    case "app":
        fmt.Println("🚀 app mode")
    case "setup":
        InitDB(db)
        SeedMasterData(db)
    case "api":
        StartAPIServer(db)
    case "help", "-h", "--help":
        printUsage()
    default:
        printUsage()
        log.Fatalf("不明なコマンドです: %s", mode)
    }
}

func printUsage() {
    fmt.Println("使用方法: /out/main [app|setup|api]")
    fmt.Println("  app   : 通常起動モード")
    fmt.Println("  setup : init + seed をまとめて実行")
    fmt.Println("  api   : JSON API サーバーを起動")
}
