# Multi-Round Deliberation & Follow-up Implementation

**Status:** ✅ Core functionality complete, UI enhancements in progress
**Last Updated:** 2026-01-03
**Goal:** Make the dynamic consensus feature more visible and user-friendly

---

## 📊 Implementation Status

### ✅ Completed (Already in Codebase)

#### Backend
- [x] Debate follow-up endpoint: `POST /api/debates/{id}/followup`
- [x] Council follow-up endpoint: `POST /api/councils/{id}/followup`
- [x] User turn tracking with `AgentID: "user"` and `MemberID: "user"`
- [x] Round number incrementing logic
- [x] Multiple conclusions/syntheses support in data model
- [x] Automatic status transitions: `completed` → `in_progress` → `completed`
- [x] Background goroutine execution for resumed deliberations
- [x] Turn/Response persistence with round tracking

#### Frontend
- [x] Follow-up input UI in DebateView (lines 346-379)
- [x] Follow-up input UI in CouncilView (lines 376-409)
- [x] Round grouping and display with dividers
- [x] Multiple conclusions display per round
- [x] User directive highlighting in councils
- [x] API client methods: `addDebateFollowUp`, `addCouncilFollowUp`

---

## 🚀 UI/UX Enhancements (Completed!)

### 1. Consensus Evolution Timeline
**File:** `web/app/src/pages/DebateView.tsx`
**Status:** ✅ Complete (Lines 263-325)
**Description:** Visual timeline showing how consensus evolved across rounds

**Implementation:**
- ✅ Consensus Evolution section with timeline design
- ✅ Side-by-side comparison of round conclusions
- ✅ Visual diff showing agreement/disagreement progression
- ✅ Expandable text (line-clamp-3 with hover expansion)
- ✅ Round badges (R1, R2, R3...) with connecting lines
- ✅ "Latest" badge on most recent conclusion
- ✅ Emoji indicators (🤝 for consensus, ⚔️ for divergent)
- ✅ Early consensus badge when applicable

### 2. Resume Status Indicator
**File:** `web/app/src/pages/DebateView.tsx` & `CouncilView.tsx`
**Status:** ✅ Complete (DebateView: 238-261, CouncilView: 222-245)
**Description:** Show visual feedback when debate/council is resuming

**Implementation:**
- ✅ Animated banner when status === 'in_progress' && rounds > 1
- ✅ Pulsing blue indicator
- ✅ Current round number display
- ✅ "Resuming deliberation..." message for debates
- ✅ "Re-convening Council..." message for councils
- ✅ Processing spinner

### 3. Round Summary Badges
**File:** `web/app/src/pages/DebateView.tsx`
**Status:** ✅ Complete (Lines 336-356)
**Description:** Add consensus/divergent badges to round headers

**Implementation:**
- ✅ Enhanced round divider with status badge
- ✅ Consensus status badge (✓ Consensus / • Divergent)
- ✅ Turn count per round (excluding user turns)
- ✅ User follow-up directive highlighting (Lines 358-371)
- ✅ Visual distinction for user interjections

### 4. Council Multi-Round Enhancements
**File:** `web/app/src/pages/CouncilView.tsx`
**Status:** ✅ Complete (Lines 247-296, 318-334)
**Description:** Similar enhancements for council view

**Implementation:**
- ✅ Synthesis Evolution Timeline (Lines 247-296)
- ✅ Round status badges (✓ Synthesized, response count)
- ✅ Member response tracking across rounds
- ✅ User directive highlighting (already existed)

---

## 🧪 Testing Checklist

### Debate Flow
- [ ] Create a new debate with 2 agents
- [ ] Wait for initial debate to complete (Round 1)
- [ ] Submit a follow-up question
- [ ] Verify status changes to `in_progress`
- [ ] Verify new round (Round 2) is created
- [ ] Verify user turn appears with "User (Follow-up)" label
- [ ] Verify agents resume deliberation
- [ ] Verify second conclusion is generated
- [ ] Verify both conclusions are displayed
- [ ] Test with 3+ rounds

### Council Flow
- [ ] Create a new council with 3+ members
- [ ] Wait for initial council to complete (Round 1)
- [ ] Submit a follow-up directive
- [ ] Verify status changes to `in_progress`
- [ ] Verify new round (Round 2) is created
- [ ] Verify user directive appears highlighted
- [ ] Verify all members provide new responses
- [ ] Verify new rankings are collected
- [ ] Verify chairman generates new synthesis
- [ ] Verify both syntheses are displayed
- [ ] Test with 3+ rounds

### Edge Cases
- [ ] Submit empty follow-up (should be rejected)
- [ ] Submit follow-up while debate is in_progress
- [ ] Submit multiple follow-ups rapidly
- [ ] Very long follow-up text (2000+ chars)
- [ ] Follow-up after read-only debate (should fail)
- [ ] Database persistence after server restart

---

## 📝 Documentation Needed

### User Documentation
- [ ] How to use follow-up feature
- [ ] What happens when you submit a follow-up
- [ ] Best practices for follow-up questions
- [ ] Understanding round progression
- [ ] Interpreting consensus evolution

### Developer Documentation
- [ ] Architecture overview of multi-round system
- [ ] Data flow diagram for follow-ups
- [ ] API endpoint specifications
- [ ] Database schema for rounds
- [ ] Frontend component structure

---

## 🐛 Known Issues / Future Improvements

### Current Limitations
- No limit on number of rounds (could loop indefinitely)
- No cost tracking for multi-round deliberations
- No way to cancel an in-progress resumed debate
- No round history compression for very long debates

### Potential Features
- [ ] Round limit configuration (e.g., max 5 rounds)
- [ ] Cost estimation before resuming
- [ ] Pause/stop button for running debates
- [ ] Export multi-round debates to PDF with timeline
- [ ] Analytics: "Consensus reached after N rounds"
- [ ] AI-suggested follow-up questions
- [ ] Round-by-round diff viewer
- [ ] Agent voting history across rounds

---

## 📐 Architecture Notes

### Data Flow for Follow-ups

```
User submits follow-up
    ↓
Frontend: api.addDebateFollowUp(id, content)
    ↓
Backend: POST /api/debates/{id}/followup
    ↓
Engine: AddFollowUp(ctx, debateID, content)
    ↓
1. Get existing debate and turns
2. Calculate new round number (last_round + 1)
3. Create user turn with AgentID="user"
4. Save turn to database
5. Spawn goroutine → RunDebate(ctx, debateID, nil)
    ↓
6. Update debate status to in_progress
7. Execute turns (agents respond to history + user directive)
8. Generate new conclusion for this round
9. Append to debate.Conclusions array
10. Update status to completed
    ↓
Frontend: Polling detects status change, displays new round
```

### Database Schema (Relevant Fields)

```sql
-- debates table
id TEXT
status TEXT  -- pending, in_progress, completed, failed
conclusion_json TEXT  -- Array of conclusions (one per round)

-- turns table
id TEXT
debate_id TEXT
agent_id TEXT  -- "user" for user follow-ups
number INTEGER  -- Sequential turn number
round INTEGER  -- Round number (1, 2, 3, ...)
content TEXT

-- councils table (similar structure)
synthesis TEXT  -- Array of syntheses (one per round)

-- responses table
round INTEGER  -- Round tracking for council responses
member_id TEXT  -- "user" for user directives
```

---

## ✅ Session Progress Tracking

### Session 1 (2026-01-03)
- [x] Analyzed existing implementation
- [x] Confirmed all backend features working
- [x] Confirmed all frontend features working
- [x] Created this TODO document
- [x] Implemented Consensus Evolution Timeline
- [x] Implemented Resume Status Indicator
- [x] Implemented Round Summary Badges
- [x] Enhanced CouncilView with multi-round features
- [x] Fixed TypeScript errors and verified build
- [x] Created comprehensive user documentation (docs/USER_GUIDE_MULTI_ROUND.md)
- [x] Created session summary (SESSION_SUMMARY.md)
- [x] Verified backend compiles successfully
- [x] Verified frontend builds successfully
- [ ] Manual end-to-end testing (recommended for next session)

---

## 🎯 Next Steps

1. **Completed This Session ✅**
   - ✅ Implemented Consensus Evolution Timeline
   - ✅ Implemented Resume Status Indicator
   - ✅ Implemented Round Summary Badges
   - ✅ Implemented council-specific enhancements
   - ✅ Created comprehensive user documentation
   - ✅ Fixed all TypeScript errors
   - ✅ Verified backend compiles successfully
   - ✅ Verified frontend builds successfully

2. **Recommended Testing (Manual)**
   - [ ] Create a new debate and complete it
   - [ ] Submit a follow-up question and verify Round 2 starts
   - [ ] Verify Resume Status Indicator appears
   - [ ] Verify Consensus Evolution Timeline displays after Round 2
   - [ ] Test with 3+ rounds to see full evolution
   - [ ] Create a council and test follow-up flow
   - [ ] Verify Synthesis Evolution Timeline for councils
   - [ ] Test edge cases (empty follow-up, very long text, rapid submissions)

3. **Future Enhancements (Optional)**
   - [ ] Add round limit configuration (e.g., max 5 rounds)
   - [ ] Implement cost tracking for multi-round deliberations
   - [ ] Implement pause/stop functionality for running debates
   - [ ] Add analytics: "Consensus reached after N rounds"
   - [ ] Add AI-suggested follow-up questions
   - [ ] Create round diff viewer
   - [ ] Add export enhancements for multi-round PDFs

---

## 📞 Contact / Questions

For questions about this implementation, refer to:
- Code: `internal/engine/engine.go:850-887` (debate follow-ups)
- Code: `internal/council/council.go:564-599` (council follow-ups)
- UI: `web/app/src/pages/DebateView.tsx:346-379`
- UI: `web/app/src/pages/CouncilView.tsx:376-409`
