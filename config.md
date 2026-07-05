---
start_date: "2026-07-13"
week_duration_days: 7
---

# Plan Configuration

Edit `start_date` to shift the entire schedule. [[Home]] recalculates all week dates automatically.

| Field | Value | Notes |
|-------|-------|-------|
| `start_date` | 2026-07-13 | W01 begins this date (Monday) |
| `week_duration_days` | 7 | Days per study week |

**W00** (Infrastructure Setup) is the week *before* start_date (Jul 6 – Jul 12 with the current config).

---

## Print the Full Schedule

```
go run tools/plan-dates.go --start 2026-07-13
```

To reschedule the entire plan to a new start date, update `start_date` above. No other files need to change.
