# PD — Home

## Current Week
<!-- Update this link every Monday when you start a new week -->
[[weeks/W01-lsm-storage]]

---

## Today's Task

| Day | Task | Duration |
|-----|------|----------|
| Mon | 📖 Read | 60 min |
| Wed | 💻 Code | 60 min |
| Thu | 🔨 Build | 60 min |
| Fri | ✍️ Blog + Reflect | 40 min |

---

## Progress

```dataview
TABLE WITHOUT ID
  file.link AS "Week",
  length(filter(file.tasks, (t) => t.completed)) + " / " + length(file.tasks) AS "Tasks"
FROM "weeks"
SORT file.name ASC
```

---

## Open Tasks This Week

```tasks
not done
path includes weeks/W01-lsm-storage
```
