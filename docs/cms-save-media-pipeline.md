# Plenti CMS save / media pipeline (reference)

> **Historical baseline:** the sections below were captured before the image-crop
> integration and contain superseded APIs and save sequencing. They must not be
> used as the current implementation contract. See [ADR D13](adr/0001-clientside-image-crop-scale-convert.md)
> and the [acceptance matrix](364-acceptance-matrix.md) for current behavior.
>
> **Current save/session contract (2026-09-13):** the CMS root owns pending-media
> cleanup on actual content-file changes and logout. Visual/Code and View/Edit
> switches preserve that state. Both Save buttons call `pendingMedia.toCommitItems`
> through `beforeSubmit`; the shared Button merges those extra items with content
> in one provider call. Only success invokes `markCommitted`. GitLab is atomic;
> Gitea writes media paths (including raw `create` and derivative `upsert`) before
> content, sequentially. Neither hook performs a separate commit.

Background map of how the Svelte CMS assembles a save and flows media values,
captured while building the client-side image-crop feature (issue #364).
Anchors are `file:line` under `defaults/core/cms/`. Companion to
[ADR 0001](adr/0001-clientside-image-crop-scale-convert.md).

## Key anchors

| Concern | Anchor | Purpose |
|---|---|---|
| Commit assembly (content) | `visual_editor.svelte:64-69` | `content.fields` → JSON commit item |
| Field binding | `visual_editor.svelte:34` | `bind:field={content.fields[label]}` |
| Schema load | `visual_editor.svelte:6-8` | `schemas[content.type]` |
| Schema → fields | `dynamic_form_input.svelte:74` | `{schema}` passed to schema-driven widgets |
| Media field NOT given schema | `dynamic_form_input.svelte:84` | `<Media …/>` lacks `{schema}` (crop plumbing gap) |
| Media trigger | `fields/media.svelte:7-11` | `swapMedia()` sets `changingMedia` |
| Media selection return | `media_grid.svelte:20-23` | `changingMedia = selected path` |
| Field value sync | `fields/media.svelte:12-16` | `$: if (changingMedia && field===originalMedia) field = changingMedia` |
| Object decomposition | `fields/fieldset.svelte:9` | `bind:field={field[key]}` (object → per-prop) |
| Upload list | `file_upload.svelte:10-18` | builds `{file, contents:dataURL}` |
| Upload commit | `file_upload.svelte:89` | `localMediaList` → `<Button commitList=…>` (eager) |
| Provider dispatch | `button.svelte` | local / gitlab / gitea (now via `providers/commit.js`) |
| Local transform | `providers/local.js:11-19` | commitList → `/postlocal` body |
| Gitlab transform | `providers/gitlab.js:44-59` | commitList → GitLab API actions |

## Save flow (content page)

```
router → admin_menu (bind:content, bind:shadowContent)
       → edit_tray → visual_editor
           schema = schemas[content.type]               // visual_editor.svelte:6-8
           per field: <DynamicFormInput bind:field={content.fields[label]} {schema} {parentKeys}>
           Save: <Button commitList={[{ file: content.filepath,
                                        contents: JSON.stringify(content.fields) }]}
                         action=… encoding="text">
       → button.svelte onSubmit → commit() → providers/{local,gitlab,gitea}
```

The page-save commit contains a single item (the content JSON). The provider
applies **one `action` + one `encoding` to the whole commit list**, which is why
media uploads commit separately (`create`/`base64`) from content (`text`).

## Media value flow

- **String field**: `dynamic_form_input.svelte:84` → `<Media bind:field>`. `swapMedia()`
  copies `field`→`changingMedia` and opens the modal; `media_grid` sets
  `changingMedia = picked`; the reactive in `media.svelte:12-16` writes it back to
  `field` (guarded by `field === originalMedia` so only the edited field updates).
- **Object field** (`{src, alt}`): routed to `Fieldset`
  (`dynamic_form_input.svelte:101`), which decomposes via
  `bind:field={field[key]}` (`fieldset.svelte:9`) and recurses into
  `DynamicFormInput` per property. The `src` property routes back to `<Media>`.
  The prototype reported nested writes not propagating; the clean fork uses the
  standard binding, so **verify in-browser before adding a workaround**.

## Upload flow

`file_upload.svelte` reads dropped/selected files to `localMediaList` items
`{file:"media/name", contents:dataURL}`, then a dedicated `<Button
commitList={localMediaList} action="create" encoding="base64">` **eagerly**
commits them and sets `changingMedia` to the uploaded path.

## Implications for the crop feature

1. **Schema plumbing**: pass `{schema} {parentKeys}` to `<Media>` at
   `dynamic_form_input.svelte:84`; read crop config via
   `parseImageOptions(schema, parentKeys)`.
2. **Deferred persistence (chosen)**: a crop produces a derivative held in the
   `pendingMedia` store (Blob, in memory). The page-save `<Button>` gets a
   `beforeSubmit` hook (`commitPendingMedia`) that commits derivatives
   (`create`/`base64`) through the **same provider** via `providers/commit.js`,
   then the existing content commit runs. Pending is cleared only after the
   whole save succeeds; cancel discards it. (Two commits, same provider —
   the provider's single-action-per-commit constraint rules out one mixed commit
   without an API change.)
3. **Field value**: keep media as a string path; the field is updated to the
   derivative's `filePath` on crop-confirm (same `changingMedia` write-back).
4. **Object fields**: ensure the derivative path reaches `field.src`; verify the
   `fieldset.svelte:9` binding propagates (it may already).
