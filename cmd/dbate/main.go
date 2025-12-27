package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/alienxp03/dbate/internal/core"
	"github.com/alienxp03/dbate/internal/engine"
	"github.com/alienxp03/dbate/internal/persona"
	"github.com/alienxp03/dbate/internal/provider"
	"github.com/alienxp03/dbate/internal/storage"
	"github.com/alienxp03/dbate/internal/style"
	"github.com/alienxp03/dbate/web/handlers"
)

var (
	dbPath string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "dbate",
	Short: "AI-powered debate tool",
	Long: `dbate is a CLI tool that orchestrates debates between AI agents.

Create debates on any topic and watch AI agents with different personas
argue, collaborate, or analyze from multiple perspectives.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Database path (default: ~/.dbate/dbate.db)")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(providersCmd)
	rootCmd.AddCommand(personasCmd)
	rootCmd.AddCommand(stylesCmd)
	rootCmd.AddCommand(serveCmd)
}

func getStorage() (storage.Storage, error) {
	path := dbPath
	if path == "" {
		path = storage.DefaultDBPath()
	}

	store, err := storage.NewSQLiteStorage(path)
	if err != nil {
		return nil, err
	}

	if err := store.Initialize(); err != nil {
		return nil, err
	}

	return store, nil
}

// new command - create and run a debate
var newCmd = &cobra.Command{
	Use:   "new [topic]",
	Short: "Start a new debate",
	Long: `Create and run a new debate on the given topic.

Examples:
  dbate new "Is AI beneficial for humanity?"
  dbate new "Best programming language for web development" --style adversarial
  dbate new "Climate change solutions" -a claude:optimist -b gemini:skeptic`,
	Args: cobra.MinimumNArgs(1),
	RunE: runNewDebate,
}

var (
	agentAFlag string
	agentBFlag string
	styleFlag  string
	turnsFlag  int
)

func init() {
	newCmd.Flags().StringVarP(&agentAFlag, "agent-a", "a", "claude:pragmatist", "Agent A config (provider:persona)")
	newCmd.Flags().StringVarP(&agentBFlag, "agent-b", "b", "claude:skeptic", "Agent B config (provider:persona)")
	newCmd.Flags().StringVarP(&styleFlag, "style", "s", "collaborative", "Debate style (adversarial, collaborative, analytical, socratic)")
	newCmd.Flags().IntVarP(&turnsFlag, "turns", "t", 5, "Number of turns per agent")
}

func parseAgentConfig(config string) (provider, persona string, err error) {
	parts := strings.SplitN(config, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid agent config: %s (expected provider:persona)", config)
	}
	return parts[0], parts[1], nil
}

func runNewDebate(cmd *cobra.Command, args []string) error {
	topic := strings.Join(args, " ")

	store, err := getStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	registry := provider.DefaultRegistry()
	eng := engine.New(store, registry)

	// Parse agent configs
	providerA, personaA, err := parseAgentConfig(agentAFlag)
	if err != nil {
		return err
	}
	providerB, personaB, err := parseAgentConfig(agentBFlag)
	if err != nil {
		return err
	}

	// Create debate
	config := core.NewDebateConfig{
		Topic:          topic,
		AgentAProvider: providerA,
		AgentAPersona:  personaA,
		AgentBProvider: providerB,
		AgentBPersona:  personaB,
		Style:          styleFlag,
		MaxTurns:       turnsFlag,
	}

	debate, err := eng.CreateDebate(cmd.Context(), config)
	if err != nil {
		return fmt.Errorf("failed to create debate: %w", err)
	}

	fmt.Printf("\n🎭 Starting Debate: %s\n", debate.Topic)
	fmt.Printf("   Style: %s | Turns: %d per agent\n", debate.Style, debate.MaxTurns)
	fmt.Printf("   Agent A: %s (%s)\n", debate.AgentA.Name, debate.AgentA.Provider)
	fmt.Printf("   Agent B: %s (%s)\n", debate.AgentB.Name, debate.AgentB.Provider)
	fmt.Printf("   ID: %s\n\n", debate.ID)
	fmt.Println(strings.Repeat("─", 60))

	// Setup signal handling
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nInterrupted. Saving debate state...")
		cancel()
	}()

	// Run debate with callback for live output
	err = eng.RunDebate(ctx, debate.ID, func(turn *core.Turn, d *core.Debate) {
		var agentName string
		if turn.AgentID == d.AgentA.ID {
			agentName = d.AgentA.Name
		} else {
			agentName = d.AgentB.Name
		}

		fmt.Printf("\n📢 Turn %d - %s\n", turn.Number, agentName)
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(turn.Content)
		fmt.Println()
	})

	if err != nil {
		if ctx.Err() != nil {
			fmt.Println("\nDebate paused. Use 'dbate show <id>' to view progress.")
			return nil
		}
		return fmt.Errorf("debate failed: %w", err)
	}

	// Show conclusion
	debate, _ = eng.GetDebate(debate.ID)
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("🏁 CONCLUSION")
	fmt.Println(strings.Repeat("═", 60))

	if debate.Conclusion != nil {
		if debate.Conclusion.Agreed {
			fmt.Println("✅ Consensus Reached!")
		} else {
			fmt.Println("❌ No Consensus")
		}
		fmt.Printf("\n%s\n", debate.Conclusion.Summary)

		if !debate.Conclusion.Agreed {
			if debate.Conclusion.AgentASummary != "" {
				fmt.Printf("\n📌 %s's Position:\n%s\n", debate.AgentA.Name, debate.Conclusion.AgentASummary)
			}
			if debate.Conclusion.AgentBSummary != "" {
				fmt.Printf("\n📌 %s's Position:\n%s\n", debate.AgentB.Name, debate.Conclusion.AgentBSummary)
			}
		}
	}

	return nil
}

// list command - list all debates
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all debates",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		eng := engine.New(store, provider.DefaultRegistry())
		debates, err := eng.ListDebates(50, 0)
		if err != nil {
			return err
		}

		if len(debates) == 0 {
			fmt.Println("No debates found. Start one with: dbate new \"Your topic\"")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTOPIC\tSTATUS\tSTYLE\tTURNS\tCREATED")
		fmt.Fprintln(w, "──\t─────\t──────\t─────\t─────\t───────")

		for _, d := range debates {
			shortID := d.ID[:8]
			shortTopic := d.Topic
			if len(shortTopic) > 40 {
				shortTopic = shortTopic[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				shortID,
				shortTopic,
				d.Status,
				d.Style,
				d.TurnCount,
				d.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()

		return nil
	},
}

// show command - show a debate
var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show debate details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		eng := engine.New(store, provider.DefaultRegistry())

		// Find debate by ID prefix
		debates, err := eng.ListDebates(100, 0)
		if err != nil {
			return err
		}

		var debateID string
		for _, d := range debates {
			if strings.HasPrefix(d.ID, args[0]) {
				debateID = d.ID
				break
			}
		}

		if debateID == "" {
			return fmt.Errorf("debate not found: %s", args[0])
		}

		debate, turns, err := eng.GetDebateWithTurns(debateID)
		if err != nil {
			return err
		}

		fmt.Printf("\n🎭 Debate: %s\n", debate.Topic)
		fmt.Printf("   ID: %s\n", debate.ID)
		fmt.Printf("   Status: %s\n", debate.Status)
		fmt.Printf("   Style: %s\n", debate.Style)
		fmt.Printf("   Agent A: %s (%s)\n", debate.AgentA.Name, debate.AgentA.Provider)
		fmt.Printf("   Agent B: %s (%s)\n", debate.AgentB.Name, debate.AgentB.Provider)
		fmt.Printf("   Created: %s\n", debate.CreatedAt.Format(time.RFC3339))
		fmt.Println()

		if len(turns) > 0 {
			fmt.Println(strings.Repeat("─", 60))
			for _, turn := range turns {
				var agentName string
				if turn.AgentID == debate.AgentA.ID {
					agentName = debate.AgentA.Name
				} else {
					agentName = debate.AgentB.Name
				}
				fmt.Printf("\n📢 Turn %d - %s\n", turn.Number, agentName)
				fmt.Println(strings.Repeat("─", 40))
				fmt.Println(turn.Content)
			}
		}

		if debate.Conclusion != nil {
			fmt.Println()
			fmt.Println(strings.Repeat("═", 60))
			fmt.Println("🏁 CONCLUSION")
			fmt.Println(strings.Repeat("═", 60))
			if debate.Conclusion.Agreed {
				fmt.Println("✅ Consensus Reached!")
			} else {
				fmt.Println("❌ No Consensus")
			}
			fmt.Printf("\n%s\n", debate.Conclusion.Summary)
		}

		return nil
	},
}

// delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a debate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		eng := engine.New(store, provider.DefaultRegistry())

		// Find by prefix
		debates, _ := eng.ListDebates(100, 0)
		var debateID string
		for _, d := range debates {
			if strings.HasPrefix(d.ID, args[0]) {
				debateID = d.ID
				break
			}
		}

		if debateID == "" {
			return fmt.Errorf("debate not found: %s", args[0])
		}

		if err := eng.DeleteDebate(debateID); err != nil {
			return err
		}

		fmt.Printf("Deleted debate: %s\n", debateID)
		return nil
	},
}

// providers command
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available AI providers",
	Run: func(cmd *cobra.Command, args []string) {
		registry := provider.DefaultRegistry()

		fmt.Println("\nAvailable Providers:")
		fmt.Println(strings.Repeat("─", 40))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDISPLAY\tSTATUS")

		for _, p := range registry.List() {
			status := "❌ Not installed"
			if p.Available() {
				status = "✅ Available"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name(), p.DisplayName(), status)
		}
		w.Flush()
	},
}

// personas command
var personasCmd = &cobra.Command{
	Use:   "personas",
	Short: "List available agent personas",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nAvailable Personas:")
		fmt.Println(strings.Repeat("─", 60))

		for _, p := range persona.DefaultPersonas() {
			fmt.Printf("\n%s (%s)\n", p.Name, p.ID)
			fmt.Printf("  %s\n", p.Description)
		}
	},
}

// styles command
var stylesCmd = &cobra.Command{
	Use:   "styles",
	Short: "List available debate styles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nAvailable Debate Styles:")
		fmt.Println(strings.Repeat("─", 60))

		for _, s := range style.DefaultStyles() {
			fmt.Printf("\n%s (%s)\n", s.Name, s.ID)
			fmt.Printf("  %s\n", s.Description)
		}
	},
}

// serve command - start web server
var (
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}
		defer store.Close()

		registry := provider.DefaultRegistry()

		// Import the handlers package
		fmt.Printf("\n🌐 Starting dbate web server on http://localhost:%d\n\n", servePort)
		fmt.Println("Available endpoints:")
		fmt.Printf("  GET  http://localhost:%d/debates     - List all debates\n", servePort)
		fmt.Printf("  GET  http://localhost:%d/new         - Create new debate form\n", servePort)
		fmt.Printf("  GET  http://localhost:%d/debates/:id - View debate\n", servePort)
		fmt.Println("\nPress Ctrl+C to stop the server")

		// Start server using the web handlers
		return startWebServer(store, registry, servePort)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8182, "Server port")
}

func startWebServer(store storage.Storage, registry *provider.Registry, port int) error {
	h := handlers.New(store, registry)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Handle shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down...")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
