# LLM Council vs conclave: Architecture Comparison

## Executive Summary

**llm-council** (by Andrej Karpathy) and **conclave** are both multi-LLM orchestration systems but solve fundamentally different problems:

- **llm-council**: Democratic consensus through peer review and synthesis
- **conclave**: Structured debate through persona-driven agents with turn-by-turn exchanges

---

## 1. Core Architecture Comparison

### llm-council: Council Consensus Model

**Philosophy**: Multiple AI models collaborate like a council to produce superior answers through peer evaluation.

**Architecture**:
```
User Question
    ↓
Stage 1: Parallel Individual Responses (all models)
    ↓
Stage 2: Peer Rankings (each model ranks others anonymously)
    ↓
Stage 3: Chairman Synthesis (one model consolidates everything)
    ↓
Final Answer
```

**Key Characteristics**:
- API-based (OpenRouter as unified gateway)
- Web-first (React + FastAPI)
- 4 council models + 1 chairman model
- Single-question focus (not conversational)
- Flat file storage (JSON conversations)
- Streaming Server-Sent Events (SSE)

### conclave: Debate Orchestration Model

**Philosophy**: Explore different perspectives through structured argumentation between agents with distinct personalities.

**Architecture**:
```
Topic + Style + Personas
    ↓
Turn-by-Turn Alternation (Agent A ↔ Agent B)
    ↓
Context-Aware Responses (history + persona + style)
    ↓
Early Consensus Detection
    ↓
Voting + Conclusion Summary
```

**Key Characteristics**:
- CLI-based (wraps existing AI CLI tools)
- Terminal + Web interfaces
- 2 agents with customizable personas
- Multi-turn conversational debate
- SQLite persistence (debates, turns, personas, styles)
- Synchronous execution with callback hooks

---

## 2. Technical Deep Dive

### Communication Model

| Aspect | llm-council | conclave |
|--------|-------------|-------|
| **API Access** | Centralized via OpenRouter | Direct CLI tool execution |
| **Models** | Any OpenRouter-supported model | claude, codex, gemini, qwen CLIs |
| **Parallelism** | Stage 1 parallel, Stage 2 parallel | Sequential turn alternation |
| **Streaming** | SSE for real-time updates | Callback hooks per turn |
| **Network** | Requires internet (API calls) | Works offline if CLIs do |

**llm-council approach**:
```python
# All models query in parallel
responses = await asyncio.gather(*[
    call_openrouter(model, prompt)
    for model in council_models
])
```

**conclave approach**:
```go
// Sequential turn execution with CLI
cmd := exec.CommandContext(ctx, "claude", prompt)
output, err := cmd.CombinedOutput()
```

**Trade-offs**:
- ✅ llm-council: Faster (parallel), consistent API interface
- ✅ conclave: No API costs, works with local models, offline-capable

---

### Storage & Persistence

| Feature | llm-council | conclave |
|---------|-------------|-------|
| **Backend** | Flat JSON files | SQLite database |
| **Schema** | Unstructured conversations | Structured (debates, turns, personas, styles) |
| **Queries** | File I/O + parsing | SQL queries with indexing |
| **History** | List directory files | Full-text search, filtering |
| **Scalability** | Limited by file system | Better for large datasets |

**llm-council storage**:
```
data/conversations/
  ├── conv_uuid1.json
  ├── conv_uuid2.json
  └── ...
```

**conclave storage**:
```sql
CREATE TABLE debates (...);
CREATE TABLE turns (...);
CREATE TABLE personas (...);
CREATE TABLE styles (...);
-- Supports queries like:
SELECT * FROM debates WHERE topic LIKE '%AI%' ORDER BY created_at DESC;
```

**Trade-offs**:
- ✅ llm-council: Simple, portable, easy to inspect
- ✅ conclave: Queryable, transactional, better at scale

---

### Orchestration Strategy

#### llm-council: Three-Stage Pipeline

**Stage 1 - Individual Responses**:
```python
async def stage1_collect_responses(content: str):
    tasks = [call_llm(model, content) for model in COUNCIL_MODELS]
    responses = await asyncio.gather(*tasks)
    return [{"model": m, "response": r} for m, r in zip(COUNCIL_MODELS, responses)]
```

**Stage 2 - Peer Rankings**:
```python
async def stage2_collect_rankings(stage1_responses):
    # Anonymize: Response A, B, C, D
    anonymized = create_anonymous_responses(stage1_responses)

    # Each model ranks others
    ranking_tasks = [
        call_llm(model, f"Rank these responses:\n{anonymized}")
        for model in COUNCIL_MODELS
    ]
    rankings = await asyncio.gather(*ranking_tasks)

    # Aggregate rankings (average positions)
    return calculate_aggregate_rankings(rankings)
```

**Stage 3 - Synthesis**:
```python
async def stage3_synthesize_final(query, stage1, stage2):
    context = f"""
    Original question: {query}

    Individual responses:
    {format_responses(stage1)}

    Peer rankings:
    {format_rankings(stage2)}

    Synthesize a final answer reflecting collective wisdom.
    """
    return await call_llm(CHAIRMAN_MODEL, context)
```

**Key Features**:
- Parallel execution in stages 1 & 2
- Anonymous peer review prevents model bias
- Aggregate rankings via average position
- Chairman role for synthesis
- No back-and-forth (single pass)

#### conclave: Turn-by-Turn Debate

**Turn Execution**:
```go
func (e *Engine) ExecuteTurn(ctx context.Context, debate *core.Debate, agent core.Agent) error {
    // 1. Fetch previous turns
    turns, _ := e.storage.GetTurns(debate.ID)

    // 2. Build prompt (persona + style + history)
    prompt := e.buildPrompt(debate, agent, turns)

    // 3. Execute provider CLI
    provider, _ := e.registry.Get(agent.Provider)
    response, err := provider.GenerateWithModel(ctx, prompt, agent.Model)

    // 4. Save turn to database
    turn := &core.Turn{
        ID:       uuid.New().String(),
        DebateID: debate.ID,
        AgentID:  agent.ID,
        Number:   len(turns) + 1,
        Content:  response,
    }
    e.storage.SaveTurn(turn)

    // 5. Trigger callback for UI updates
    if e.turnCallback != nil {
        e.turnCallback(debate, turn)
    }

    return nil
}
```

**Alternation Pattern**:
```go
// Determine starting agent (random but deterministic per debate ID)
currentAgent := determineStartingAgent(debate)

for turnNumber := 0; turnNumber < debate.MaxTurns * 2; turnNumber++ {
    // Execute turn for current agent
    ExecuteTurn(ctx, debate, currentAgent)

    // Early consensus check (every 2 turns after turn 4)
    if turnNumber >= 4 && turnNumber % 2 == 1 {
        if checkConsensus(debate) {
            break  // End debate early
        }
    }

    // Switch agents
    currentAgent = (currentAgent == AgentA) ? AgentB : AgentA
}
```

**Key Features**:
- Sequential alternation (Agent A → Agent B → A → B...)
- Context accumulation (each turn sees full history)
- Early consensus detection (keyword matching + verification)
- Persona-driven prompts (optimist, skeptic, etc.)
- Style-driven templates (adversarial, collaborative, etc.)

**Trade-offs**:
- ✅ llm-council: Comprehensive evaluation, reduces groupthink
- ✅ conclave: Deeper exploration through dialogue, more nuanced interaction

---

### Prompt Engineering

#### llm-council: Role-Based Prompts

**Ranking Prompt**:
```
You are evaluating multiple responses to: "{question}"

Here are the anonymized responses:

Response A: {response_a}
Response B: {response_b}
Response C: {response_c}
Response D: {response_d}

Analyze each response for accuracy and insight.
Provide detailed critique.

FINAL RANKING:
1. Response [letter]
2. Response [letter]
...
```

**Synthesis Prompt**:
```
As the Chairman, synthesize these responses and rankings into one comprehensive answer.

Original question: {question}
Individual responses: {all_responses}
Peer rankings: {aggregated_rankings}

Provide a balanced final answer reflecting collective wisdom.
```

#### conclave: Template-Based Dynamic Prompts

**Prompt Formula**:
```
[Persona System Prompt] + [Style Template with Variables]
```

**Example - Pragmatist in Collaborative Style**:
```
System Prompt (Persona):
You are a pragmatist who focuses on practical, implementable solutions.
You prioritize resource constraints and proven approaches.

Style Template (Collaborative Response):
Topic: {{.Topic}}
You are {{.AgentName}} discussing with {{.OtherAgentName}}.

{{.OtherAgentName}} said:
{{.PreviousArgument}}

Build on their ideas while adding your perspective.
Focus on finding common ground and synthesizing viewpoints.
Turn {{.TurnNumber}} of {{.MaxTurns}}.
```

**Variable Substitution**:
```go
type TemplateData struct {
    Topic             string
    AgentName         string
    OtherAgentName    string
    PreviousArgument  string
    DebateHistory     string
    TurnNumber        int
    MaxTurns          int
    IsQuestioner      bool  // For Socratic style
}
```

**Trade-offs**:
- ✅ llm-council: Focused, structured evaluation
- ✅ conclave: Highly customizable, supports complex debate dynamics

---

## 3. Feature Comparison Matrix

| Feature | llm-council | conclave |
|---------|-------------|-------|
| **Multi-model support** | ✅ Via OpenRouter | ✅ Via CLI tools |
| **Parallel execution** | ✅ Stages 1 & 2 | ❌ Sequential turns |
| **Peer evaluation** | ✅ Anonymous rankings | ❌ (but has voting) |
| **Turn-by-turn debate** | ❌ Single-pass | ✅ Multi-turn with history |
| **Personas** | ❌ Fixed council roles | ✅ 6 builtin + custom |
| **Debate styles** | ❌ One approach | ✅ 4 builtin + custom |
| **Early consensus** | ❌ Always 3 stages | ✅ Keyword + verification |
| **Export** | ❌ JSON only | ✅ Markdown, PDF, JSON |
| **CLI interface** | ❌ Web only | ✅ Full-featured CLI |
| **Web interface** | ✅ React SPA | ✅ Go templates + HTMX |
| **Step mode** | ❌ Automatic only | ✅ Manual turn-by-turn |
| **Read-only lock** | ❌ | ✅ Prevent modifications |
| **Streaming** | ✅ SSE | ❌ (callback hooks) |
| **Offline capable** | ❌ Requires API | ✅ If CLI tools support |
| **Cost** | 💰 API calls | Free (CLI-based) |

---

## 4. What's Better in Each Project

### llm-council Advantages

#### 1. **Parallel Processing**
```python
# All council models respond simultaneously
responses = await asyncio.gather(*[
    call_model(model, query) for model in council
])
```
- ⚡ Faster overall execution (stages run in parallel)
- 🎯 Better for time-sensitive queries

#### 2. **Peer Review System**
- **Anonymous evaluation** prevents bias
- **Aggregate rankings** identify best answers mathematically
- **Democratic process** reduces single-model weaknesses
- **Quality scoring** via average rank position

#### 3. **Chairman Synthesis**
- Dedicated model for final answer
- Considers both responses AND rankings
- Creates coherent narrative from multiple perspectives
- Reflects "collective wisdom"

#### 4. **Unified API Interface**
- Single OpenRouter API for all models
- Consistent response format
- Easier error handling
- No CLI dependency management

#### 5. **Real-time Streaming**
```python
async def stream_response():
    yield f"data: {json.dumps(stage1_result)}\n\n"
    yield f"data: {json.dumps(stage2_result)}\n\n"
    yield f"data: {json.dumps(stage3_result)}\n\n"
```
- Progressive UI updates as stages complete
- Better user experience for long operations

---

### conclave Advantages

#### 1. **Rich Persona System**
- 6 builtin personas with distinct thinking styles
- Custom persona support via SQLite
- Persona-specific system prompts shape behavior
- **Example**: Skeptic vs Optimist creates natural tension

#### 2. **Flexible Debate Styles**
- 4 builtin styles (adversarial, collaborative, analytical, socratic)
- Custom styles via template system
- Style determines interaction pattern
- **Example**: Socratic uses question/answer alternation

#### 3. **Context Accumulation**
```go
// Each turn sees full debate history
prompt := buildPrompt(
    persona,
    style,
    allPreviousTurns,  // Cumulative context
    currentTurn
)
```
- Agents learn from entire discussion
- Arguments evolve based on previous exchanges
- Natural progression of ideas

#### 4. **Early Consensus Detection**
```go
func checkConsensus(turns []Turn) bool {
    // Keyword matching
    keywords := []string{"i agree", "consensus", "common ground"}
    if containsKeywords(turns, keywords) {
        // Verification prompt
        response := askAgent("Have you reached consensus?")
        return strings.Contains(response, "YES")
    }
    return false
}
```
- Saves compute when agents agree early
- Natural conversation ending
- Practical efficiency gain

#### 5. **CLI-First Design**
- No API costs (uses local CLI tools)
- Works offline (if CLIs support it)
- Integrates with existing workflows
- **Example**: `conclave new "topic" | tee debate.txt`

#### 6. **SQLite Persistence**
- Queryable history
- Full-text search
- Transactional integrity
- Better scalability
- **Example**: `SELECT * FROM debates WHERE topic LIKE '%AI%'`

#### 7. **Export Flexibility**
```bash
conclave export <id> markdown  # Clean Markdown
conclave export <id> pdf       # Formatted PDF
conclave export <id> json      # Machine-readable
```
- Multiple output formats
- Share debates easily
- Archive for reference

#### 8. **Step-by-Step Mode**
```bash
conclave new "topic" --step
# Press Enter to advance turn
# Type 'q' to quit
```
- Manual control over debate flow
- Better for live demonstrations
- Pedagogical value

#### 9. **Dual Interface**
- Full CLI for automation/scripting
- Web UI for exploration
- Best of both worlds

---

## 5. Architectural Improvements conclave Can Learn

### 1. **Parallel Initial Responses** ⭐⭐⭐

**Current conclave**: Sequential turns (A → B → A → B...)
**llm-council**: All models respond simultaneously in Stage 1

**Potential Enhancement**:
```go
// New feature: Parallel opening statements
func (e *Engine) CollectOpeningStatements(ctx context.Context, debate *core.Debate) ([]Turn, error) {
    var wg sync.WaitGroup
    results := make(chan Turn, 2)

    // Agent A and B respond in parallel to initial topic
    wg.Add(2)
    go func() {
        defer wg.Done()
        turn, _ := e.generateOpening(ctx, debate.AgentA, debate.Topic)
        results <- turn
    }()
    go func() {
        defer wg.Done()
        turn, _ := e.generateOpening(ctx, debate.AgentB, debate.Topic)
        results <- turn
    }()

    wg.Wait()
    close(results)

    turns := []Turn{}
    for turn := range results {
        turns = append(turns, turn)
    }
    return turns, nil
}
```

**Benefits**:
- ⚡ Faster initial turn (2x speedup)
- 🎯 Both agents form independent initial positions
- 📊 Clearer starting divergence

**Implementation Complexity**: Low
**Impact**: Medium-High

---

### 2. **Peer Evaluation / Self-Reflection** ⭐⭐⭐⭐

**Current conclave**: Voting happens only at conclusion
**llm-council**: Continuous peer evaluation and ranking

**Potential Enhancement**:
```go
// After debate, ask each agent to evaluate the OTHER agent's arguments
type Evaluation struct {
    EvaluatorID string
    TargetID    string
    Strengths   []string
    Weaknesses  []string
    Rating      int  // 1-10
}

func (e *Engine) CollectPeerEvaluations(ctx context.Context, debate *core.Debate) []Evaluation {
    // Agent A evaluates Agent B's arguments
    evalA := e.evaluateArguments(ctx, debate.AgentA, getAgentBTurns(debate))

    // Agent B evaluates Agent A's arguments
    evalB := e.evaluateArguments(ctx, debate.AgentB, getAgentATurns(debate))

    return []Evaluation{evalA, evalB}
}

func (e *Engine) evaluateArguments(ctx context.Context, evaluator Agent, targetTurns []Turn) Evaluation {
    prompt := fmt.Sprintf(`
    Evaluate the following arguments objectively:

    %s

    Provide:
    1. Key strengths (3-5 points)
    2. Key weaknesses (3-5 points)
    3. Overall rating (1-10)

    Format:
    STRENGTHS:
    - Point 1
    - Point 2

    WEAKNESSES:
    - Point 1
    - Point 2

    RATING: X/10
    `, formatTurns(targetTurns))

    response, _ := e.executeProvider(ctx, evaluator, prompt)
    return parseEvaluation(response)
}
```

**Benefits**:
- 🎓 Deeper analysis of arguments
- 🤝 Identifies genuine weaknesses and strengths
- 📈 More informative conclusions
- 🧠 Meta-cognitive layer (thinking about thinking)

**Use Cases**:
- Teaching critical thinking
- Identifying logical fallacies
- Improving argument quality

**Implementation Complexity**: Medium
**Impact**: High

---

### 3. **Anonymous Turn Presentation** ⭐⭐

**Current conclave**: Agents know opponent identity
**llm-council**: Anonymous peer review prevents bias

**Potential Enhancement**:
```go
// Option to anonymize agent identities in prompts
type DebateConfig struct {
    AnonymousMode bool  // Hide agent identities
}

func (e *Engine) buildPrompt(debate *core.Debate, agent Agent, turns []Turn) string {
    if debate.AnonymousMode {
        // Replace agent names with "Speaker A" and "Speaker B"
        history := anonymizeHistory(turns)
    } else {
        history := formatHistory(turns)
    }
    // ... rest of prompt building
}
```

**Benefits**:
- 🎭 Reduces persona-based bias
- 🧪 Tests pure argument quality
- 📊 Research tool for studying bias

**Implementation Complexity**: Low
**Impact**: Low-Medium

---

### 4. **Streaming / Progressive Updates** ⭐⭐⭐

**Current conclave**: Callback hooks per turn (not web-streaming)
**llm-council**: Server-Sent Events for real-time updates

**Potential Enhancement**:
```go
// Add SSE endpoint for web interface
func (h *Handler) StreamDebate(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, _ := w.(http.Flusher)

    // Create debate with callback
    debate := engine.CreateDebate(...)

    engine.SetTurnCallback(func(d *core.Debate, turn *core.Turn) {
        // Send SSE event for each turn
        data := map[string]interface{}{
            "type": "turn",
            "turn": turn,
        }
        fmt.Fprintf(w, "data: %s\n\n", toJSON(data))
        flusher.Flush()
    })

    engine.RunDebate(ctx, debate)

    // Send completion event
    fmt.Fprintf(w, "data: {\"type\":\"complete\"}\n\n")
    flusher.Flush()
}
```

**Benefits**:
- ⚡ Better web UX (see progress in real-time)
- 🎯 Works with long-running debates
- 📱 Modern web app feel

**Implementation Complexity**: Medium
**Impact**: High (for web users)

---

### 5. **Aggregate Ranking System** ⭐⭐⭐⭐

**Current conclave**: Simple binary voting (AGREE/DISAGREE)
**llm-council**: Numerical rankings with averaging

**Potential Enhancement**:
```go
type RankedArgument struct {
    TurnID    string
    AgentID   string
    Content   string
    Rankings  []int     // Each agent's rank for this turn
    AvgRank   float64   // Average rank
    Votes     int       // Number of votes
}

func (e *Engine) RankTurns(ctx context.Context, debate *core.Debate) []RankedArgument {
    turns := e.storage.GetTurns(debate.ID)

    // Ask each agent to rank ALL turns (including their own)
    rankingsA := e.getRankings(ctx, debate.AgentA, turns)
    rankingsB := e.getRankings(ctx, debate.AgentB, turns)

    // Calculate aggregate rankings
    rankedArgs := []RankedArgument{}
    for _, turn := range turns {
        rankA := rankingsA[turn.ID]
        rankB := rankingsB[turn.ID]

        rankedArgs = append(rankedArgs, RankedArgument{
            TurnID:   turn.ID,
            AgentID:  turn.AgentID,
            Content:  turn.Content,
            Rankings: []int{rankA, rankB},
            AvgRank:  float64(rankA + rankB) / 2.0,
            Votes:    2,
        })
    }

    // Sort by average rank (lower is better)
    sort.Slice(rankedArgs, func(i, j int) func() bool {
        return rankedArgs[i].AvgRank < rankedArgs[j].AvgRank
    })

    return rankedArgs
}

func (e *Engine) getRankings(ctx context.Context, agent Agent, turns []Turn) map[string]int {
    prompt := fmt.Sprintf(`
    Rank the following arguments from BEST (1) to WORST (%d):

    %s

    Format your response as:
    RANKINGS:
    1. Turn [id]
    2. Turn [id]
    ...
    `, len(turns), formatTurnsForRanking(turns))

    response, _ := e.executeProvider(ctx, agent, prompt)
    return parseRankings(response, turns)
}
```

**Benefits**:
- 📊 Identifies strongest arguments objectively
- 🏆 Highlights key turning points in debate
- 📈 Quantitative analysis of debate quality
- 🎯 Better conclusion summaries

**Implementation Complexity**: Medium-High
**Impact**: High

---

### 6. **Multi-Agent Synthesis** ⭐⭐

**Current conclave**: Two agents vote, then neutral summary
**llm-council**: Chairman model synthesizes all inputs

**Potential Enhancement**:
```go
type SynthesisConfig struct {
    SynthesizerProvider string  // e.g., "claude/opus"
    SynthesizerPersona  string  // e.g., "neutral_synthesizer"
}

func (e *Engine) GenerateSynthesis(ctx context.Context, debate *core.Debate) string {
    // Get all debate content
    turns := e.storage.GetTurns(debate.ID)
    evaluations := e.CollectPeerEvaluations(ctx, debate)
    rankings := e.RankTurns(ctx, debate)

    // Use a third-party model to synthesize
    synthesizerPrompt := fmt.Sprintf(`
    You are a neutral synthesizer reviewing a debate.

    Topic: %s

    Debate transcript:
    %s

    Peer evaluations:
    %s

    Argument rankings:
    %s

    Provide a comprehensive synthesis that:
    1. Summarizes key points from both sides
    2. Identifies areas of agreement and disagreement
    3. Highlights the strongest arguments (by ranking)
    4. Offers a balanced conclusion
    `, debate.Topic, formatTurns(turns), formatEvaluations(evaluations), formatRankings(rankings))

    // Use a different model for synthesis (e.g., Opus for quality)
    synthesizer := e.registry.Get(debate.SynthesisConfig.SynthesizerProvider)
    synthesis, _ := synthesizer.Generate(ctx, synthesizerPrompt)

    return synthesis
}
```

**Benefits**:
- 🎯 More objective conclusions
- 🧠 Leverages different model strengths
- 📊 Higher quality final summaries

**Implementation Complexity**: Low-Medium
**Impact**: Medium

---

## 6. Implementation Priority Matrix

| Enhancement | Complexity | Impact | Priority | Effort |
|-------------|-----------|---------|----------|--------|
| **Parallel opening statements** | Low | Medium-High | 🔥 High | 1-2 days |
| **Streaming SSE** | Medium | High | 🔥 High | 2-3 days |
| **Peer evaluations** | Medium | High | 🔥 High | 3-4 days |
| **Aggregate ranking** | Medium-High | High | 🔥 High | 4-5 days |
| **Multi-agent synthesis** | Low-Medium | Medium | 🟡 Medium | 2-3 days |
| **Anonymous mode** | Low | Low-Medium | 🟢 Low | 1 day |

---

## 7. Recommended Roadmap

### Phase 1: Quick Wins (1 week)
1. ✅ **Parallel opening statements** - Immediate 2x speedup on first turn
2. ✅ **Streaming SSE for web** - Better UX without architectural changes

### Phase 2: Core Enhancements (2 weeks)
3. ✅ **Peer evaluation system** - Add post-debate analysis
4. ✅ **Aggregate ranking** - Identify best arguments mathematically

### Phase 3: Advanced Features (1 week)
5. ✅ **Multi-agent synthesis** - Third-party conclusion generator
6. ✅ **Anonymous mode** - Research feature for bias studies

---

## 8. Key Takeaways

### What conclave does better:
- ✅ **Richer interaction model** (turn-by-turn with context)
- ✅ **Persona & style flexibility** (customizable debate formats)
- ✅ **CLI-first design** (cost-effective, offline-capable)
- ✅ **Persistence** (SQLite with queryable history)
- ✅ **Export options** (Markdown, PDF, JSON)

### What llm-council does better:
- ✅ **Parallel execution** (faster overall)
- ✅ **Peer review system** (anonymous, democratic)
- ✅ **Aggregate rankings** (quantitative quality assessment)
- ✅ **Chairman synthesis** (dedicated model for conclusions)
- ✅ **Streaming UX** (real-time web updates)

### Best hybrid approach:
```
conclave's architecture
+ llm-council's parallelism (opening statements)
+ llm-council's peer review (post-debate analysis)
+ llm-council's ranking system (argument quality scoring)
+ llm-council's streaming (web UX)
= Next-generation debate platform
```

---

## 9. Conclusion

**llm-council** and **conclave** solve different problems:

- **llm-council**: Optimized for **quality** (peer review, synthesis, democratic consensus)
- **conclave**: Optimized for **exploration** (personas, styles, turn-by-turn dialogue)

The most valuable improvements for conclave are:
1. **Peer evaluation system** - Adds depth without changing core model
2. **Aggregate rankings** - Quantifies argument quality
3. **Parallel opening statements** - Performance boost
4. **Streaming SSE** - Modern web experience

These enhancements preserve conclave's unique strengths (personas, styles, CLI) while adopting llm-council's best ideas (peer review, rankings, synthesis).
