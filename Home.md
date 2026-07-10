# PD Home

## Schedule

```dataviewjs
const config = dv.page("config");
if (!config || !config.start_date) {
  dv.paragraph("⚠️ config.md not found or missing start_date.");
} else {
  // config.start_date may be a Dataview date object or a string; handle both
  const raw = config.start_date;
  const w01Start = typeof raw === "string"
    ? dv.luxon.DateTime.fromISO(raw)
    : dv.luxon.DateTime.fromISO(raw.toISODate ? raw.toISODate() : String(raw));

  const today = dv.luxon.DateTime.now().startOf("day");

  const pages = dv.pages('"weeks"')
    .filter(p => p.week_number !== undefined)
    .sort(p => p.week_number);

  let currentLink = null;

  const rows = pages.map(p => {
    const offset = p.week_number === 0 ? -7 : (p.week_number - 1) * 7;
    const weekStart = w01Start.plus({ days: offset });
    const weekEnd   = weekStart.plus({ days: 6 });

    const isCurrent = today >= weekStart && today <= weekEnd;
    const isPast    = today > weekEnd;

    const completed = p.file.tasks.where(t => t.completed).length;
    const total     = p.file.tasks.length;
    const progress  = total > 0 ? `${completed} / ${total}` : "-";
    const status    = isCurrent ? "👉 now"
                    : isPast    ? "✓ past"
                    :             (p.status ?? "-");

    if (isCurrent) currentLink = p.file.link;

    return [p.file.link, weekStart.toFormat("MMM d"), weekEnd.toFormat("MMM d"), progress, status];
  });

  if (currentLink) dv.paragraph("**Current week:** " + currentLink);

  dv.table(["Week", "Start", "End", "Tasks", "Status"], rows);
}
```

---

## Open Tasks This Week

```dataviewjs
const config = dv.page("config");
if (config && config.start_date) {
  const raw = config.start_date;
  const w01Start = typeof raw === "string"
    ? dv.luxon.DateTime.fromISO(raw)
    : dv.luxon.DateTime.fromISO(raw.toISODate ? raw.toISODate() : String(raw));

  const today = dv.luxon.DateTime.now().startOf("day");

  const current = dv.pages('"weeks"')
    .filter(p => p.week_number !== undefined)
    .sort(p => p.week_number)
    .find(p => {
      const offset = p.week_number === 0 ? -7 : (p.week_number - 1) * 7;
      const s = w01Start.plus({ days: offset });
      const e = s.plus({ days: 6 });
      return today >= s && today <= e;
    });

  if (current) {
    const open = current.file.tasks.where(t => !t.completed);
    dv.paragraph("**" + current.file.name + "**: " + open.length + " open tasks");
    dv.taskList(open, false);
  } else {
    dv.paragraph("No active week. Check `start_date` in [[config]].");
  }
}
```
