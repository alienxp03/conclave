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

	"github.com/alienxp03/dbate/internal/config"
	"github.com/alienxp03/dbate/internal/core"
	"github.com/alienxp03/dbate/internal/engine"
	"github.com/alienxp03/dbate/internal/export"
	"github.com/alienxp03/dbate/internal/persona"
	"github.com/alienxp03/dbate/internal/provider"
	"github.com/alienxp03/dbate/internal/storage"
	"github.com/alienxp03/dbate/internal/style"
	"github.com/alienxp03/dbate/web/handlers"
)

var (
	dbPath    string
	cfgPath   string
	appConfig *config.Config
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		var err error
		if cfgPath != "" {
			appConfig, err = config.LoadFrom(cfgPath)
		} else {
			appConfig, err = config.Load()
		}
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Database path (default: ~/.dbate/dbate.db)")
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Config file path (default: ~/.dbate/config.yaml)")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(providersCmd)
	rootCmd.AddCommand(personasCmd)
	rootCmd.AddCommand(stylesCmd)
	rootCmd.AddCommand(configCmd)
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

func getRegistry() *provider.Registry {
	if appConfig != nil {
		return provider.RegistryFromConfig(appConfig)
	}
	return provider.DefaultRegistry()
}

// ============================================================================
// NEW COMMAND
// ============================================================================

var newCmd = &cobra.Command{
	Use:   "new [topic]",
	Short: "Start a new debate",
	Long: `Create and run a new debate on the given topic.

Examples:
  dbate new "Is AI beneficial for humanity?"
  dbate new "Best programming language" --style adversarial
  dbate new "Climate change" -a claude:optimist -b gemini:skeptic
  dbate new "Tech trends" -a claude/sonnet:analyst -b qwen:visionary --step`,
	Args: cobra.MinimumNArgs(1),
	RunE: runNewDebate,
}

var (
	agentAFlag   string
	agentBFlag   string
	styleFlag    string
	turnsFlag    int
	stepModeFlag bool
)

func init() {
	newCmd.Flags().StringVarP(&agentAFlag, "agent-a", "a", "claude:pragmatist", "Agent A (provider[/model]:persona)")
	newCmd.Flags().StringVarP(&agentBFlag, "agent-b", "b", "claude:skeptic", "Agent B (provider[/model]:persona)")
	newCmd.Flags().StringVarP(&styleFlag, "style", "s", "collaborative", "Debate style")
	newCmd.Flags().IntVarP(&turnsFlag, "turns", "t", 5, "Turns per agent")
	newCmd.Flags().BoolVar(&stepModeFlag, "step", false, "Step-by-step mode (execute one turn at a time)")
}

// parseAgentConfig parses "provider[/model]:persona" format
func parseAgentConfig(cfg string) (prov, model, pers string, err error) {
	parts := strings.SplitN(cfg, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid agent config: %s (expected provider[/model]:persona)", cfg)
	}

	providerPart := parts[0]
	pers = parts[1]

	// Check for model
	if strings.Contains(providerPart, "/") {
		provParts := strings.SplitN(providerPart, "/", 2)
		prov = provParts[0]
		model = provParts[1]
	} else {
		prov = providerPart
	}

	return prov, model, pers, nil
}

func runNewDebate(cmd *cobra.Command, args []string) error {
	topic := strings.Join(args, " ")

	store, err := getStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	registry := getRegistry()
	eng := engine.New(store, registry)

	// Parse agent configs
	providerA, modelA, personaA, err := parseAgentConfig(agentAFlag)
	if err != nil {
		return err
	}
	providerB, modelB, personaB, err := parseAgentConfig(agentBFlag)
	if err != nil {
		return err
	}

	// Set mode
	mode := core.ModeAutomatic
	if stepModeFlag {
		mode = core.ModeTurnByTurn
	}

	// Create debate
	debateConfig := core.NewDebateConfig{
		Topic:          topic,
		AgentAProvider: providerA,
		AgentAModel:    modelA,
		AgentAPersona:  personaA,
		AgentBProvider: providerB,
		AgentBModel:    modelB,
		AgentBPersona:  personaB,
		Style:          styleFlag,
		Mode:           mode,
		MaxTurns:       turnsFlag,
	}

	debate, err := eng.CreateDebate(cmd.Context(), debateConfig)
	if err != nil {
		return fmt.Errorf("failed to create debate: %w", err)
	}

	fmt.Printf("\n🎭 Debate: %s\n", debate.Topic)
	fmt.Printf("   Style: %s | Turns: %d per agent | Mode: %s\n", debate.Style, debate.MaxTurns, debate.Mode)
	fmt.Printf("   Agent A: %s (%s", debate.AgentA.Name, debate.AgentA.Provider)
	if debate.AgentA.Model != "" {
		fmt.Printf("/%s", debate.AgentA.Model)
	}
	fmt.Println(")")
	fmt.Printf("   Agent B: %s (%s", debate.AgentB.Name, debate.AgentB.Provider)
	if debate.AgentB.Model != "" {
		fmt.Printf("/%s", debate.AgentB.Model)
	}
	fmt.Println(")")
	fmt.Printf("   ID: %s\n\n", debate.ID)
	fmt.Println(strings.Repeat("─", 60))

	// Step mode - let user control
	if stepModeFlag {
		return runStepMode(cmd.Context(), eng, debate)
	}

	// Auto mode - run full debate
	return runAutoMode(cmd.Context(), eng, debate)
}

func runAutoMode(ctx context.Context, eng *engine.Engine, debate *core.Debate) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nInterrupted. Saving debate state...")
		cancel()
	}()

	err := eng.RunDebate(ctx, debate.ID, func(turn *core.Turn, d *core.Debate) {
		agentName := getAgentName(d, turn.AgentID)
		fmt.Printf("\n📢 Turn %d - %s\n", turn.Number, agentName)
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(turn.Content)
		fmt.Println()
	})

	if err != nil {
		if ctx.Err() != nil {
			fmt.Println("\nDebate paused. Resume with: dbate show " + debate.ID[:8])
			return nil
		}
		return fmt.Errorf("debate failed: %w", err)
	}

	return showConclusion(eng, debate.ID)
}

func runStepMode(ctx context.Context, eng *engine.Engine, debate *core.Debate) error {
	fmt.Println("\n📋 Step-by-step mode. Commands:")
	fmt.Println("   [Enter] - Execute next turn")
	fmt.Println("   [q]     - Quit and save")
	fmt.Println()

	totalTurns := debate.MaxTurns * 2
	currentTurn := 0

	for currentTurn < totalTurns {
		fmt.Printf("Turn %d/%d - Press Enter to continue (q to quit): ", currentTurn+1, totalTurns)

		var input string
		fmt.Scanln(&input)

		if strings.ToLower(strings.TrimSpace(input)) == "q" {
			fmt.Println("\nDebate paused. Resume with: dbate show " + debate.ID[:8])
			return nil
		}

		turn, err := eng.ExecuteNextTurn(ctx, debate.ID)
		if err != nil {
			return fmt.Errorf("failed to execute turn: %w", err)
		}

		debate, _, _ = eng.GetDebateWithTurns(debate.ID)
		agentName := getAgentName(debate, turn.AgentID)

		fmt.Printf("\n📢 Turn %d - %s\n", turn.Number, agentName)
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(turn.Content)
		fmt.Println()

		currentTurn++
	}

	return showConclusion(eng, debate.ID)
}

func showConclusion(eng *engine.Engine, debateID string) error {
	debate, _ := eng.GetDebate(debateID)
	if debate == nil || debate.Conclusion == nil {
		return nil
	}

	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("🏁 CONCLUSION")
	fmt.Println(strings.Repeat("═", 60))

	// Show votes
	if debate.Conclusion.AgentAVote != nil {
		voteIcon := "❌"
		if debate.Conclusion.AgentAVote.Agrees {
			voteIcon = "✅"
		}
		fmt.Printf("\n%s %s votes: %s\n", voteIcon, debate.AgentA.Name,
			map[bool]string{true: "AGREE", false: "DISAGREE"}[debate.Conclusion.AgentAVote.Agrees])
	}
	if debate.Conclusion.AgentBVote != nil {
		voteIcon := "❌"
		if debate.Conclusion.AgentBVote.Agrees {
			voteIcon = "✅"
		}
		fmt.Printf("%s %s votes: %s\n", voteIcon, debate.AgentB.Name,
			map[bool]string{true: "AGREE", false: "DISAGREE"}[debate.Conclusion.AgentBVote.Agrees])
	}

	fmt.Println()
	if debate.Conclusion.Agreed {
		if debate.Conclusion.EarlyConsensus {
			fmt.Println("🤝 Consensus Reached Early!")
		} else {
			fmt.Println("🤝 Consensus Reached!")
		}
	} else {
		fmt.Println("⚔️  No Consensus")
	}

	fmt.Printf("\n%s\n", debate.Conclusion.Summary)

	if !debate.Conclusion.Agreed {
		if debate.Conclusion.AgentASummary != "" {
			fmt.Printf("\n📌 %s:\n%s\n", debate.AgentA.Name, debate.Conclusion.AgentASummary)
		}
		if debate.Conclusion.AgentBSummary != "" {
			fmt.Printf("\n📌 %s:\n%s\n", debate.AgentB.Name, debate.Conclusion.AgentBSummary)
		}
	}

	return nil
}

func getAgentName(debate *core.Debate, agentID string) string {
	if agentID == debate.AgentA.ID {
		return debate.AgentA.Name
	}
	return debate.AgentB.Name
}

// ============================================================================
// LIST COMMAND
// ============================================================================

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all debates",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		eng := engine.New(store, getRegistry())
		debates, err := eng.ListDebates(50, 0)
		if err != nil {
			return err
		}

		if len(debates) == 0 {
			fmt.Println("No debates found. Start one with: dbate new \"Your topic\"")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTOPIC\tSTATUS\tMODE\tTURNS\tLOCK\tCREATED")
		fmt.Fprintln(w, "──\t─────\t──────\t────\t─────\t────\t───────")

		for _, d := range debates {
			shortID := d.ID[:8]
			shortTopic := d.Topic
			if len(shortTopic) > 35 {
				shortTopic = shortTopic[:32] + "..."
			}
			lock := ""
			if d.ReadOnly {
				lock = "🔒"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				shortID,
				shortTopic,
				d.Status,
				d.Mode,
				d.TurnCount,
				lock,
				d.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()

		return nil
	},
}

// ============================================================================
// SHOW COMMAND
// ============================================================================

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

		eng := engine.New(store, getRegistry())
		debateID, err := findDebateByPrefix(eng, args[0])
		if err != nil {
			return err
		}

		debate, turns, err := eng.GetDebateWithTurns(debateID)
		if err != nil {
			return err
		}

		fmt.Printf("\n🎭 Debate: %s\n", debate.Topic)
		fmt.Printf("   ID: %s\n", debate.ID)
		fmt.Printf("   Status: %s\n", debate.Status)
		fmt.Printf("   Style: %s | Mode: %s\n", debate.Style, debate.Mode)
		if debate.ReadOnly {
			fmt.Println("   🔒 Read-only")
		}
		fmt.Printf("   Agent A: %s (%s)\n", debate.AgentA.Name, debate.AgentA.Provider)
		fmt.Printf("   Agent B: %s (%s)\n", debate.AgentB.Name, debate.AgentB.Provider)
		fmt.Printf("   Created: %s\n", debate.CreatedAt.Format(time.RFC3339))
		fmt.Println()

		if len(turns) > 0 {
			fmt.Println(strings.Repeat("─", 60))
			for _, turn := range turns {
				agentName := getAgentName(debate, turn.AgentID)
				fmt.Printf("\n📢 Turn %d - %s\n", turn.Number, agentName)
				fmt.Println(strings.Repeat("─", 40))
				fmt.Println(turn.Content)
			}
		}

		if debate.Conclusion != nil {
			showConclusion(eng, debate.ID)
		}

		return nil
	},
}

// ============================================================================
// DELETE COMMAND
// ============================================================================

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

		eng := engine.New(store, getRegistry())
		debateID, err := findDebateByPrefix(eng, args[0])
		if err != nil {
			return err
		}

		if err := eng.DeleteDebate(debateID); err != nil {
			return err
		}

		fmt.Printf("Deleted debate: %s\n", debateID)
		return nil
	},
}

// ============================================================================
// EXPORT COMMAND
// ============================================================================

var exportCmd = &cobra.Command{
	Use:   "export [id] [format]",
	Short: "Export debate to file",
	Long: `Export a debate to markdown, PDF, or JSON.

Examples:
  dbate export abc123 markdown
  dbate export abc123 pdf
  dbate export abc123 json -o debate.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		eng := engine.New(store, getRegistry())
		debateID, err := findDebateByPrefix(eng, args[0])
		if err != nil {
			return err
		}

		debate, turns, err := eng.GetDebateWithTurns(debateID)
		if err != nil {
			return err
		}

		format := export.Format(strings.ToLower(args[1]))
		exporter, err := export.GetExporter(format)
		if err != nil {
			return err
		}

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = export.GenerateFilename(debate, exporter.FileExtension())
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		if err := exporter.Export(debate, turns, file); err != nil {
			return fmt.Errorf("failed to export: %w", err)
		}

		fmt.Printf("Exported to: %s\n", outputPath)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringP("output", "o", "", "Output file path")
}

// ============================================================================
// LOCK COMMAND
// ============================================================================

var lockCmd = &cobra.Command{
	Use:   "lock [id]",
	Short: "Toggle read-only lock on a debate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore, ok := store.(*storage.SQLiteStorage)
		if !ok {
			return fmt.Errorf("lock not supported for this storage type")
		}

		eng := engine.New(store, getRegistry())
		debateID, err := findDebateByPrefix(eng, args[0])
		if err != nil {
			return err
		}

		debate, err := eng.GetDebate(debateID)
		if err != nil {
			return err
		}

		newState := !debate.ReadOnly
		if err := sqlStore.SetReadOnly(debateID, newState); err != nil {
			return err
		}

		if newState {
			fmt.Printf("🔒 Locked debate: %s\n", debateID[:8])
		} else {
			fmt.Printf("🔓 Unlocked debate: %s\n", debateID[:8])
		}
		return nil
	},
}

// ============================================================================
// PROVIDERS COMMAND
// ============================================================================

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available AI providers",
	Run: func(cmd *cobra.Command, args []string) {
		registry := getRegistry()

		fmt.Println("\nAvailable Providers:")
		fmt.Println(strings.Repeat("─", 50))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDISPLAY\tMODELS\tSTATUS")

		for _, p := range registry.List() {
			status := "❌ Not installed"
			if p.Available() {
				status = "✅ Available"
			}
			models := strings.Join(p.Models(), ", ")
			if len(models) > 30 {
				models = models[:27] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name(), p.DisplayName(), models, status)
		}
		w.Flush()
	},
}

// ============================================================================
// PERSONAS COMMAND
// ============================================================================

var personasCmd = &cobra.Command{
	Use:   "persona",
	Short: "Manage agent personas",
	Aliases: []string{"personas"},
}

var personaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all personas",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)

		fmt.Println("\nBuilt-in Personas:")
		fmt.Println(strings.Repeat("─", 60))
		for _, p := range persona.DefaultPersonas() {
			fmt.Printf("\n%s (%s) [builtin]\n", p.Name, p.ID)
			fmt.Printf("  %s\n", p.Description)
		}

		customs, err := sqlStore.ListPersonas(false)
		if err != nil {
			return err
		}

		if len(customs) > 0 {
			fmt.Println("\nCustom Personas:")
			fmt.Println(strings.Repeat("─", 60))
			for _, p := range customs {
				fmt.Printf("\n%s (%s)\n", p.Name, p.ID)
				fmt.Printf("  %s\n", p.Description)
			}
		}

		return nil
	},
}

var personaShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show persona details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Check builtin first
		if p := persona.Get(id); p != nil {
			fmt.Printf("\nPersona: %s (%s) [builtin]\n", p.Name, p.ID)
			fmt.Printf("Description: %s\n", p.Description)
			fmt.Println("\nSystem Prompt:")
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(p.SystemPrompt)
			return nil
		}

		// Check custom
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)
		p, err := sqlStore.GetPersona(id)
		if err != nil {
			return err
		}
		if p == nil {
			return fmt.Errorf("persona not found: %s", id)
		}

		fmt.Printf("\nPersona: %s (%s)\n", p.Name, p.ID)
		fmt.Printf("Description: %s\n", p.Description)
		fmt.Printf("Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Println("\nSystem Prompt:")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(p.SystemPrompt)
		return nil
	},
}

var personaCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new persona",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")
		prompt, _ := cmd.Flags().GetString("prompt")

		if id == "" || name == "" || prompt == "" {
			return fmt.Errorf("--id, --name, and --prompt are required")
		}

		// Check for conflict with builtin
		if persona.Get(id) != nil {
			return fmt.Errorf("cannot use ID '%s': conflicts with builtin persona", id)
		}

		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)

		p := &storage.Persona{
			ID:           id,
			Name:         name,
			Description:  desc,
			SystemPrompt: prompt,
		}

		if err := sqlStore.CreatePersona(p); err != nil {
			return err
		}

		fmt.Printf("Created persona: %s (%s)\n", name, id)
		return nil
	},
}

var personaDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a custom persona",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Prevent deletion of builtins
		if persona.Get(id) != nil {
			return fmt.Errorf("cannot delete builtin persona: %s", id)
		}

		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)
		if err := sqlStore.DeletePersona(id); err != nil {
			return err
		}

		fmt.Printf("Deleted persona: %s\n", id)
		return nil
	},
}

func init() {
	personaCreateCmd.Flags().String("id", "", "Persona ID (required)")
	personaCreateCmd.Flags().String("name", "", "Persona name (required)")
	personaCreateCmd.Flags().String("description", "", "Persona description")
	personaCreateCmd.Flags().String("prompt", "", "System prompt (required)")

	personasCmd.AddCommand(personaListCmd)
	personasCmd.AddCommand(personaShowCmd)
	personasCmd.AddCommand(personaCreateCmd)
	personasCmd.AddCommand(personaDeleteCmd)
}

// ============================================================================
// STYLES COMMAND
// ============================================================================

var stylesCmd = &cobra.Command{
	Use:   "style",
	Short: "Manage debate styles",
	Aliases: []string{"styles"},
}

var styleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all styles",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)

		fmt.Println("\nBuilt-in Styles:")
		fmt.Println(strings.Repeat("─", 60))
		for _, s := range style.DefaultStyles() {
			fmt.Printf("\n%s (%s) [builtin]\n", s.Name, s.ID)
			fmt.Printf("  %s\n", s.Description)
		}

		customs, err := sqlStore.ListStyles(false)
		if err != nil {
			return err
		}

		if len(customs) > 0 {
			fmt.Println("\nCustom Styles:")
			fmt.Println(strings.Repeat("─", 60))
			for _, s := range customs {
				fmt.Printf("\n%s (%s)\n", s.Name, s.ID)
				fmt.Printf("  %s\n", s.Description)
			}
		}

		return nil
	},
}

var styleShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show style details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Check builtin first
		if s := style.Get(id); s != nil {
			fmt.Printf("\nStyle: %s (%s) [builtin]\n", s.Name, s.ID)
			fmt.Printf("Description: %s\n", s.Description)
			fmt.Println("\nOpening Prompt:")
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(s.OpeningPrompt)
			fmt.Println("\nResponse Prompt:")
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(s.ResponsePrompt)
			fmt.Println("\nConclusion Prompt:")
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(s.ConclusionPrompt)
			return nil
		}

		// Check custom
		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)
		s, err := sqlStore.GetStyle(id)
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("style not found: %s", id)
		}

		fmt.Printf("\nStyle: %s (%s)\n", s.Name, s.ID)
		fmt.Printf("Description: %s\n", s.Description)
		fmt.Printf("Created: %s\n", s.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Println("\nOpening Prompt:")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(s.OpeningPrompt)
		fmt.Println("\nResponse Prompt:")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(s.ResponsePrompt)
		fmt.Println("\nConclusion Prompt:")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println(s.ConclusionPrompt)
		return nil
	},
}

var styleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new style",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")
		opening, _ := cmd.Flags().GetString("opening")
		response, _ := cmd.Flags().GetString("response")
		conclusion, _ := cmd.Flags().GetString("conclusion")

		if id == "" || name == "" {
			return fmt.Errorf("--id and --name are required")
		}
		if opening == "" || response == "" || conclusion == "" {
			return fmt.Errorf("--opening, --response, and --conclusion prompts are required")
		}

		// Check for conflict with builtin
		if style.Get(id) != nil {
			return fmt.Errorf("cannot use ID '%s': conflicts with builtin style", id)
		}

		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)

		s := &storage.Style{
			ID:               id,
			Name:             name,
			Description:      desc,
			OpeningPrompt:    opening,
			ResponsePrompt:   response,
			ConclusionPrompt: conclusion,
		}

		if err := sqlStore.CreateStyle(s); err != nil {
			return err
		}

		fmt.Printf("Created style: %s (%s)\n", name, id)
		return nil
	},
}

var styleDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a custom style",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Prevent deletion of builtins
		if style.Get(id) != nil {
			return fmt.Errorf("cannot delete builtin style: %s", id)
		}

		store, err := getStorage()
		if err != nil {
			return err
		}
		defer store.Close()

		sqlStore := store.(*storage.SQLiteStorage)
		if err := sqlStore.DeleteStyle(id); err != nil {
			return err
		}

		fmt.Printf("Deleted style: %s\n", id)
		return nil
	},
}

func init() {
	styleCreateCmd.Flags().String("id", "", "Style ID (required)")
	styleCreateCmd.Flags().String("name", "", "Style name (required)")
	styleCreateCmd.Flags().String("description", "", "Style description")
	styleCreateCmd.Flags().String("opening", "", "Opening prompt template (required)")
	styleCreateCmd.Flags().String("response", "", "Response prompt template (required)")
	styleCreateCmd.Flags().String("conclusion", "", "Conclusion prompt template (required)")

	stylesCmd.AddCommand(styleListCmd)
	stylesCmd.AddCommand(styleShowCmd)
	stylesCmd.AddCommand(styleCreateCmd)
	stylesCmd.AddCommand(styleDeleteCmd)
}

// ============================================================================
// CONFIG COMMAND
// ============================================================================

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Config file: %s\n\n", config.DefaultConfigPath())

		if appConfig != nil {
			fmt.Println("Current settings:")
			fmt.Printf("  Default style: %s\n", appConfig.Defaults.Style)
			fmt.Printf("  Default turns: %d\n", appConfig.Defaults.MaxTurns)
			fmt.Printf("  Default provider: %s\n", appConfig.Defaults.Provider)
			fmt.Println("\nProviders:")
			for name, p := range appConfig.Providers {
				status := "disabled"
				if p.Enabled {
					status = "enabled"
				}
				fmt.Printf("  %s: %s (timeout: %s)\n", name, status, p.Timeout)
			}
		}
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create example config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultConfigPath()
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}

		example := config.GenerateExample()
		if err := os.MkdirAll(strings.TrimSuffix(path, "/config.yaml"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(example), 0644); err != nil {
			return err
		}

		fmt.Printf("Created config at: %s\n", path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
}

// ============================================================================
// SERVE COMMAND
// ============================================================================

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

		registry := getRegistry()

		fmt.Printf("\n🌐 Starting dbate web server on http://localhost:%d\n\n", servePort)
		fmt.Println("Available endpoints:")
		fmt.Printf("  GET  http://localhost:%d/debates     - List all debates\n", servePort)
		fmt.Printf("  GET  http://localhost:%d/new         - Create new debate form\n", servePort)
		fmt.Printf("  GET  http://localhost:%d/debates/:id - View debate\n", servePort)
		fmt.Println("\nPress Ctrl+C to stop the server")

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

// ============================================================================
// HELPERS
// ============================================================================

func findDebateByPrefix(eng *engine.Engine, prefix string) (string, error) {
	debates, _ := eng.ListDebates(100, 0)
	for _, d := range debates {
		if strings.HasPrefix(d.ID, prefix) {
			return d.ID, nil
		}
	}
	return "", fmt.Errorf("debate not found: %s", prefix)
}
