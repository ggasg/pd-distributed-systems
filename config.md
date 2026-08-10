---
start_date: "2026-08-17"
week_duration_days: 7
---

# Plan Configuration

Edit `start_date` to shift the entire schedule. [[Home]] recalculates all week dates automatically.

| Field | Value | Notes |
|-------|-------|-------|
| `start_date` | 2026-08-17 | W01 begins this date (Monday) |
| `week_duration_days` | 7 | Days per study week |

**W00** (Infrastructure Setup) is the unit *before* start_date, and is budgeted at about 7 hours.

**These dates are a running order, not deadlines.** Each unit is scoped to about 10 hours of work rather than to seven calendar days. If a unit takes you ten days, that is the plan working as intended, not you falling behind.

---

## Print the Full Schedule

```
go run tools/plan-dates.go --start 2026-08-17
```

To reschedule the entire plan to a new start date, update `start_date` above. No other files need to change.
