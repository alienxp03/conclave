# Architecture Decision: Debate Styles vs Peer Review (2-agent vs N-agent)

## The Question

Should dbate:
1. **Keep debate styles** (optimized for 2 agents)
2. **Switch to peer review** (N-agent council like llm-council)
3. **Support both** (hybrid system)

## Current State (2-Agent Debate Styles)

### How it works:
```
Topic: "Is AI beneficial?"
├─ Agent A (Optimist)
├─ Agent B (Skeptic)
├─ Style: Adversarial
└─ Flow:
    Turn 1: A makes opening statement
    Turn 2: B counters
    Turn 3: A responds to B
    Turn 4: B responds to A
    ...
    Conclusion: Both vote + summary
```

### Strengths:
- ✅ **Deep dialogue** - agents build on each other's arguments
- ✅ **Personas create tension** - optimist vs skeptic is natural
- ✅ **Context accumulation** - each turn references previous turns
- ✅ **Multiple interaction patterns** - adversarial, collaborative, socratic, analytical
- ✅ **Natural conversation flow** - feels like a real debate
- ✅ **Unique positioning** - no one else does this

### Limitations:
- ❌ **Only 2 agents** - doesn't scale to 3+ naturally
- ❌ **Linear/sequential** - slower than parallel
- ❌ **No peer evaluation** - agents don't step back
- ❌ **Binary voting** - just A and B agree/disagree

### Can debate styles extend to 3+ agents?

**Adversarial with 3 agents:**
```
Turn 1: Agent A
Turn 2: Agent B
Turn 3: Agent C
Turn 4: Agent A (responding to... B? C? both?)
```
🤔 **Problem**: Who responds to whom? Round-robin feels artificial.

**Collaborative with 3 agents:**
```
Turn 1: Agent A proposes idea
Turn 2: Agent B builds on it
Turn 3: Agent C adds perspective
Turn 4: Agent A synthesizes B and C
```
✅ **Works better**: Collaborative could work with 3+ agents

**Socratic with 3 agents:**
```
Questioner: Agent A
Responders: Agents B, C, D (all answer questions)
```
⚠️ **Awkward**: Socratic is inherently 1-on-1

**Verdict**: Most debate styles are **fundamentally 2-agent** patterns.

---

## Alternative: N-Agent Peer Review (llm-council style)

### How it works:
```
Topic: "Is AI beneficial?"
├─ Agents: A, B, C, D (4 models)
├─ No style (one pattern)
└─ Flow:
    Stage 1: All respond independently (parallel)
    Stage 2: All rank each other's responses (parallel)
    Stage 3: Chairman synthesizes
```

### Strengths:
- ✅ **Scales to N agents** - 4, 8, 12 models work the same
- ✅ **Parallel execution** - faster
- ✅ **Democratic** - peer voting eliminates bias
- ✅ **Quantitative** - aggregate rankings show consensus
- ✅ **Diverse perspectives** - more models = broader coverage

### Limitations:
- ❌ **No dialogue** - agents don't respond to each other
- ❌ **Personas less meaningful** - all just vote on quality
- ❌ **One pattern** - no adversarial/collaborative/socratic distinction
- ❌ **Less differentiated** - basically copying llm-council
- ❌ **Shallow interaction** - single-pass, no depth

---

## Key Insight: These Serve Different Use Cases

### Debate Styles (2-agent) → **Exploration & Depth**
**User intent**: "I want to explore different perspectives on this topic"

**Example use cases**:
- Philosophy discussions (ethics, values)
- Design decisions (exploring trade-offs)
- Learning (Socratic method)
- Creative ideation (visionary vs pragmatist)

**Value**: **Quality of dialogue**, not quantity of perspectives

### Peer Review (N-agent) → **Consensus & Breadth**
**User intent**: "What's the most accurate/best answer?"

**Example use cases**:
- Factual questions (multiple models verify)
- Code review (many reviewers spot issues)
- Risk assessment (diverse expertise)
- Quality filtering (best answer wins)

**Value**: **Breadth of perspectives**, democratic consensus

---

## Option 1: Keep Debate Styles (Stay Differentiated)

### Recommendation:
Focus on **2-agent debates** but add **post-debate analysis**:

```bash
# During: Deep 2-agent dialogue
dbate new "topic" -a claude:optimist -b gemini:skeptic -s adversarial

# After: N-agent analysis
dbate analyze <id> --reviewers claude/opus,gemini/pro,qwen/max
```

**Implementation**:
```go
// Core: 2-agent debate (keep current engine)
type Debate struct {
    AgentA Agent
    AgentB Agent
    Style  string  // adversarial, collaborative, etc.
    Turns  []Turn
}

// New: Post-debate analysis with N reviewers
type Analysis struct {
    DebateID   string
    Reviewers  []Agent  // 1-N models review the debate
    Rankings   map[string][]int  // Reviewer -> turn rankings
    Synthesis  string   // One model synthesizes
}

func (e *Engine) AnalyzeDebate(debate *Debate, reviewers []Agent) (*Analysis, error) {
    // 1. Get all turns from debate
    turns := e.storage.GetTurns(debate.ID)

    // 2. Parallel: Ask each reviewer to rank turns
    rankings := e.collectRankings(turns, reviewers)

    // 3. Generate synthesis from rankings
    synthesis := e.synthesizeAnalysis(debate, rankings)

    return &Analysis{
        DebateID:  debate.ID,
        Reviewers: reviewers,
        Rankings:  rankings,
        Synthesis: synthesis,
    }
}
```

**Benefits**:
- ✅ Keep unique debate styles (differentiation)
- ✅ Add N-agent perspectives (breadth)
- ✅ Best of both worlds
- ✅ Backward compatible

**CLI**:
```bash
# Standard 2-agent debate
dbate new "topic" -a claude:skeptic -b gemini:optimist

# Add 3rd party analysis
dbate analyze <id> --add-reviewer qwen/max

# Or analyze with full council
dbate analyze <id> --reviewers claude/opus,gemini/pro,qwen/max,codex/gpt4
```

---

## Option 2: Switch to Peer Review (Copy llm-council)

### What this means:
- Remove debate styles (adversarial, collaborative, etc.)
- Remove 2-agent dialogue
- Implement 3-stage council (responses → rankings → synthesis)

**Pros**:
- ✅ Scales to N agents easily
- ✅ Faster (parallel execution)
- ✅ Proven pattern (llm-council works)

**Cons**:
- ❌ **Lose unique value** - become a CLI version of llm-council
- ❌ **Lose dialogue depth** - no turn-by-turn conversation
- ❌ **Lose personas** - optimist/skeptic less meaningful in peer review
- ❌ **Lose styles** - adversarial/collaborative/socratic don't apply

**Verdict**: ❌ **Don't recommend**
You'd be competing directly with llm-council in their strength while abandoning your differentiators.

---

## Option 3: Support Both (Hybrid Architecture)

### Two modes:

**Mode 1: Debate (2-agent, turn-by-turn)**
```bash
dbate new "topic" --mode debate \
  -a claude:optimist \
  -b gemini:skeptic \
  -s adversarial
```

**Mode 2: Council (N-agent, peer review)**
```bash
dbate new "topic" --mode council \
  --models claude/opus,gemini/pro,qwen/max,codex/gpt4 \
  --chairman claude/opus
```

### Implementation:
```go
type DebateMode string

const (
    ModeDebate  DebateMode = "debate"   // 2-agent dialogue
    ModeCouncil DebateMode = "council"  // N-agent peer review
)

type DebateConfig struct {
    Mode DebateMode

    // Debate mode fields
    AgentA Agent
    AgentB Agent
    Style  string

    // Council mode fields
    CouncilMembers []Agent
    Chairman       Agent
}
```

**Pros**:
- ✅ Maximum flexibility
- ✅ Appeal to both use cases
- ✅ Users choose the right tool

**Cons**:
- ❌ **Complex codebase** - two different engines
- ❌ **Maintenance burden** - 2x features to maintain
- ❌ **User confusion** - when to use which mode?
- ❌ **Diluted focus** - neither mode gets polish

**Verdict**: ⚠️ **Possible but risky**
Only if you have strong demand for both patterns.

---

## Option 4: Hybrid - Best of Both (Recommended)

Keep **debate as core**, add **council as analysis layer**:

### Architecture:
```
┌─────────────────────────────────────┐
│         Debate (2-agent)            │
│  • Turn-by-turn dialogue            │
│  • Personas (optimist, skeptic)     │
│  • Styles (adversarial, etc.)       │
│  • Deep exploration                 │
└──────────────┬──────────────────────┘
               │ outputs debate transcript
               ▼
┌─────────────────────────────────────┐
│      Analysis (N-agent, optional)   │
│  • Multiple reviewers read debate   │
│  • Rank turns by quality            │
│  • Synthesize findings              │
│  • Generate meta-insights           │
└─────────────────────────────────────┘
```

### Example workflow:
```bash
# 1. Run debate (core feature)
dbate new "Should we use microservices?" \
  -a claude:pragmatist \
  -b gemini:visionary \
  -s analytical

# 2. Optionally add council review
dbate review <id> \
  --reviewers claude/opus,gemini/pro,qwen/max \
  --chairman claude/opus

# Output:
# COUNCIL REVIEW
# ============
#
# Turn Rankings (by quality):
# 1. Turn 4 (Pragmatist): "Resource constraints..." - Avg rank: 1.3
# 2. Turn 2 (Visionary): "Scalability benefits..." - Avg rank: 2.0
# 3. Turn 6 (Pragmatist): "Migration costs..." - Avg rank: 2.7
#
# Reviewer Consensus:
# - Claude Opus: Pragmatist had stronger evidence
# - Gemini Pro: Both made valid points, tie
# - Qwen Max: Visionary overlooked operational complexity
#
# Chairman Synthesis:
# "The debate revealed tension between long-term scalability (Visionary)
#  and short-term pragmatism (Pragmatist). The strongest arguments focused
#  on resource constraints and migration costs. Recommendation: Hybrid
#  approach - monolith with service boundaries."
```

### Benefits:
- ✅ **Core identity**: Debate-focused (unique)
- ✅ **Optional depth**: Council review when needed
- ✅ **Layered value**: Basic (2 agents) → Advanced (N reviewers)
- ✅ **Backward compatible**: Existing debates work as-is
- ✅ **Clear positioning**: "Deep debates + optional council review"

### CLI commands:
```bash
# Core debate commands (unchanged)
dbate new <topic>
dbate list
dbate show <id>

# New analysis commands
dbate review <id> [options]         # Add council review
dbate review <id> --show            # Show existing review
dbate rank <id>                     # Show turn rankings
```

---

## Decision Matrix

| Criteria | Keep Debate | Switch to Council | Support Both | Hybrid (Debate + Review) |
|----------|-------------|-------------------|--------------|--------------------------|
| **Differentiation** | ⭐⭐⭐⭐ | ⭐ | ⭐⭐ | ⭐⭐⭐⭐ |
| **Dialogue depth** | ⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Scales to N agents** | ⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Complexity** | ⭐⭐⭐⭐ (simple) | ⭐⭐⭐ | ⭐ (complex) | ⭐⭐⭐ |
| **Maintenance** | ⭐⭐⭐⭐ (easy) | ⭐⭐⭐ | ⭐ (hard) | ⭐⭐⭐ |
| **Backward compat** | ⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Learning curve** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| **Unique value** | ⭐⭐⭐⭐ | ⭐ | ⭐⭐ | ⭐⭐⭐⭐ |

---

## Recommendation: Hybrid (Debate + Review)

### Why:
1. **Keep your moat**: Debate styles with personas are unique
2. **Add breadth**: N-agent review when users want it
3. **Backward compatible**: Existing users unaffected
4. **Clear value ladder**:
   - Basic: 2-agent debate
   - Advanced: + council review
   - Expert: Custom reviewers per debate

### Implementation plan:

**Phase 1: Core review command (1 week)**
```bash
dbate review <id> --reviewers claude,gemini,qwen
```
- Read debate transcript
- Ask each reviewer to rank turns
- Calculate aggregate rankings
- Save to database

**Phase 2: Synthesis (1 week)**
```bash
dbate review <id> --chairman claude/opus
```
- Use chairman model to synthesize
- Consider rankings + debate content
- Generate meta-insights

**Phase 3: CLI polish (3 days)**
```bash
dbate rank <id>              # Show turn rankings table
dbate review <id> --show     # Display existing review
dbate export <id> --with-review  # Include in PDF
```

**Phase 4: Web UI (optional, 3 days)**
- Show review results in debate view
- Button: "Add Council Review"
- Visual: Turn rankings chart

---

## Alternative: Start Small

If you want to test the concept first:

**Minimal viable feature:**
```bash
# Just add ranking, no full council
dbate rank <id> --reviewer claude/opus

# Output:
# Turn Rankings by claude/opus:
# 1. Turn 4 - "Strong evidence for..." (9/10)
# 2. Turn 2 - "Good point but..." (7/10)
# 3. Turn 6 - "Weak reasoning..." (5/10)
```

Then evolve based on user feedback:
- If popular → add multi-reviewer aggregation
- If not → keep it simple

---

## My Recommendation

**Go with Hybrid (Debate + Review)**:

```bash
# Your core product (unique)
dbate new "topic" -a claude:skeptic -b gemini:optimist -s adversarial

# Optional enhancement (breadth)
dbate review <debate-id> --reviewers claude,gemini,qwen,codex

# Export everything
dbate export <debate-id> pdf --with-review
```

This gives you:
- ✅ **Unique positioning**: Deep debate dialogues (core)
- ✅ **Competitive feature**: Multi-model review (optional)
- ✅ **Clear migration path**: Existing debates work as-is
- ✅ **Value differentiation**: You're not just another llm-council clone

**Keep debate styles. Add council review as a post-processing layer.**

Does this framing help? What's your gut reaction?
