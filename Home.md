# Distributed Data Systems

[Study order](README.md) · [Setup](SETUP.md) · [Resources](RESOURCES.md) · [Measurements](MEASUREMENTS.md)

```dataviewjs
const raw = dv.page("config")?.start_date;
const iso = typeof raw === "string" ? raw : raw?.toISODate?.();
const start = iso ? dv.luxon.DateTime.fromISO(iso) : null;
const pages = dv.pages('"weeks"').where(p => Number.isInteger(p.week_number));
const core = pages.where(p => p.track === "core").sort(p => p.week_number);
const extension = pages.where(p => p.track === "extension").sort(p => p.week_number);
const progress = p => {
  const tasks = p.file.tasks;
  return tasks.length ? `${tasks.where(t => t.completed).length} / ${tasks.length}` : "–";
};
dv.header(2, "Core");
dv.table(["Slot", "Unit", "Suggested start", "Sessions"], core.map(p => [
  p.week_number === 0 ? "Setup" : p.week_number,
  dv.fileLink(p.file.path, false, p.title ?? p.file.name),
  start?.isValid ? start.plus({ days: (p.week_number - 1) * 7 }).toFormat("MMM d, yyyy") : "–",
  progress(p),
]));
const next = core.find(p => p.file.tasks.some(t => !t.completed));
if (next) {
  dv.header(3, "First unfinished core unit");
  dv.paragraph(next.file.link);
  dv.taskList(next.file.tasks.where(t => !t.completed), false);
}
dv.header(2, "ML extension");
dv.table(["Slot", "Unit", "Sessions"], extension.map(p => [p.week_number, dv.fileLink(p.file.path, false, p.title ?? p.file.name), progress(p)]));
```
