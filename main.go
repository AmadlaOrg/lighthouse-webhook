package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AmadlaOrg/lighthouse-webhook/webhook"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	appName = "lighthouse-webhook"
	version = "1.0.0"
)

var (
	infoOutputFlag string
	infoHeryFlag   bool
)

var rootCmd = &cobra.Command{
	Use:     appName,
	Short:   "Lighthouse plugin for sending notifications via HTTP webhook",
	Version: version,
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show plugin metadata",
	Run: func(cmd *cobra.Command, args []string) {
		metadata := map[string]any{
			"name":        appName,
			"version":     version,
			"channel":     "webhook",
			"description": "Sends notifications via HTTP webhook",
		}
		if err := writeInfoOutput(os.Stdout, infoOutputFlag, infoHeryFlag, metadata); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding metadata: %v\n", err)
			os.Exit(1)
		}
	},
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a notification via webhook",
	Long:  "Reads alert JSON from stdin and sends it to the configured webhook URL.",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot read stdin: %v\n", err)
			os.Exit(2)
		}

		sender := webhook.New()
		if err := sender.Send(data); err != nil {
			result := map[string]any{
				"delivered": false,
				"channel":   "webhook",
				"message":   err.Error(),
			}
			enc := json.NewEncoder(os.Stdout)
			enc.Encode(result)
			os.Exit(1)
		}

		result := map[string]any{
			"delivered": true,
			"channel":   "webhook",
			"message":   "Webhook delivered successfully",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	},
}

func init() {
	infoCmd.Flags().StringVarP(&infoOutputFlag, "output", "o", "table", "Output format: table, json, yaml")
	infoCmd.Flags().BoolVar(&infoHeryFlag, "hery", false, "Wrap output in HERY envelope (_type, _body)")
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(sendCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

type heryEnvelope struct {
	Type string `json:"_type" yaml:"_type"`
	Body any    `json:"_body" yaml:"_body"`
}

func writeInfoOutput(w io.Writer, format string, hery bool, data map[string]any) error {
	if hery {
		return writeHeryOutput(w, format, data)
	}

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml":
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(bytes))
		return err
	default:
		table := tablewriter.NewWriter(w)
		table.Header("Field", "Value")
		table.Append("Name", fmt.Sprint(data["name"]))
		table.Append("Version", fmt.Sprint(data["version"]))
		if ch, ok := data["channel"]; ok {
			table.Append("Channel", fmt.Sprint(ch))
		}
		if eng, ok := data["engine"]; ok {
			table.Append("Engine", fmt.Sprint(eng))
		}
		table.Append("Description", fmt.Sprint(data["description"]))
		if supports, ok := data["supports"].([]string); ok {
			table.Append("Supports", strings.Join(supports, "\n"))
		}
		table.Render()
		return nil
	}
}

func writeHeryOutput(w io.Writer, format string, data map[string]any) error {
	envelope := heryEnvelope{
		Type: "amadla.org/entity/tools/info@v1.0.0",
		Body: data,
	}

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(envelope)
	case "table":
		fmt.Fprintf(w, "_type: %s\n\n", envelope.Type)
		table := tablewriter.NewWriter(w)
		table.Header("Field", "Value")
		table.Append("Name", fmt.Sprint(data["name"]))
		table.Append("Version", fmt.Sprint(data["version"]))
		if ch, ok := data["channel"]; ok {
			table.Append("Channel", fmt.Sprint(ch))
		}
		if eng, ok := data["engine"]; ok {
			table.Append("Engine", fmt.Sprint(eng))
		}
		table.Append("Description", fmt.Sprint(data["description"]))
		if supports, ok := data["supports"].([]string); ok {
			table.Append("Supports", strings.Join(supports, "\n"))
		}
		table.Render()
		return nil
	default:
		bytes, err := yaml.Marshal(envelope)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(bytes))
		return err
	}
}
