package main

import (
        "context"
        "flag"
        "fmt"
        "log/slog"
        "os"
        "os/signal"
        "syscall"
        "text/tabwriter"
        "vmuser/cmd"
        "vmuser/config"
)

func main() {
        configFile := flag.String("config", "vmuser.toml", "Path to the configuration file")
        tui := flag.Bool("tui", false, "Run TUI")
        addReport := flag.String("add-report", "", "Path to the report file to add")
        getReport := flag.Int64("get-report", -1, "ID of the report to retrieve")
        listReports := flag.Bool("list-reports", false, "List all reports")

        flag.Parse()

        appContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
        defer stop()

        cfg := config.GetVMUserConfig(*configFile)

        // Handle report commands
        if *addReport != "" {
                if err := cmd.AddReport(appContext, cfg, *addReport); err != nil {
                        slog.Error("Error adding report", "error", err, "file", *addReport)
                        os.Exit(1)
                }
                return
        }

        if *getReport >= 0 {
                report, err := cmd.GetReportByID(appContext, cfg, *getReport)
                if err != nil {
                        slog.Error("Error getting report", "error", err, "id", *getReport)
                        os.Exit(1)
                }
                w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
                cmd.DisplayReport(w, report)
                w.Flush()
                return
        }

        if *listReports {
                reports, err := cmd.ListAllReports(appContext, cfg)
                if err != nil {
                        slog.Error("Error listing reports", "error", err)
                        os.Exit(1)
                }
                w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
                fmt.Fprintln(w, "ID\tFilename\tCreated At")
                fmt.Fprintln(w, "---\t--------\t----------")
                for _, r := range reports {
                        fmt.Fprintf(w, "%d\t%s\t%s\n",
                                r.ID,
                                r.Filename,
                                r.CreatedAt.Format("2006-01-02 15:04:05"))
                }
                w.Flush()
                return
        }

        if *tui {
                if err := cmd.TUI(appContext, cfg); err != nil {
                        slog.Error("Error running application", "error", err)
                        os.Exit(1)
                }
        }

        if err := cmd.Server(appContext, cfg); err != nil {
                slog.Error("Error running application", "error", err)
                os.Exit(1)
        }
}
