# #364 — browser regression + remote-provider smoke evidence

Companion to [`364-acceptance-matrix.md`](364-acceptance-matrix.md). Records the
manual integration evidence (local fixture browser run) and the live-remote
smoke expectations for the parts the automated suites cannot reach.

## A. Local browser regression (built binary serving `crop-fixture`)

Environment: `plenti serve` on the fork binary, local mode (`env.local=true` →
CMS editable → `/postlocal` writes to disk). Chrome via automation.

| Scenario | Result |
|---|---|
| Standalone optimise (whole image) | `big-photo-01e60f1b…-2048x1229.webp` 388KB→64KB, valid WebP |
| Standalone crop | zoom→336×336 sub-crop; `big-photo-98ab9e6b…-336x336.webp`, distinct hash from the whole-image derivative |
| Mixed multi-file queue (JPEG, PNG, PDF, SVG, GIF) | deterministic order; 2 image derivatives + 3 byte-identical passthrough; GIF stayed animated (89a) |
| Field ordinary/no-schema upload | path assigned directly (no crop) — covered by Slice 2 |
| Field `crop:true` upload → auto-return | gateway optimise-only (no crop toggle) → media modal closes → **one** field crop modal opens; on Apply, `pendingMedia` stages the derivative, field value = path |
| Field `crop:true` SUCCESS end-to-end | Library-pick `perry.webp` → field crop → Apply → `pendingMedia` blob preview → **page Save** flushed `media/perry-500x300.webp` to disk + content JSON |
| Cancel current | Skip on image 1/3 advanced to 2/3; skipped item produced no saved file |
| Cancel all | modal closed, queue cleared, **0** `cancel-test-*` files on disk |
| Refresh persistence | after reload + reopen tray, `hero_string` field renders `media/perry-500x300.webp` (naturalWidth 500×300); value survived in `generated/content.js`; derivative in `public/media/` |
| Content JSON purity | `data:image` count = **0** after every operation |
| Object URLs | **0** blob-backed `<img>` after modals close / reload; **0** console errors across the whole session |

### Forced failure paths (the two Slice-2 ownership boundaries)

| Path | Injection | Result |
|---|---|---|
| **Provider failure before canonical save** | one-shot `/postlocal` → 500 on a field-launched eager commit | field value unchanged (`media/perry.webp` on disk), `media[]` unchanged, **no file written**, modal stays actionable, error = "Save failed (500): simulated provider failure" (an upload/provider failure, not a field failure). Exactly 1 postlocal call (no silent retry). |
| **Canonical save OK, field processing fails after** | gateway commit succeeds; one-shot `getContext('2d')→null` on the *field* crop transform | canonical asset **persisted** (`fail-path2-01e60f1b…-2048x1229.webp` on disk, valid WebP), media modal **already closed**, error surfaced in the field crop UI as a transform/field failure ("source image … invalid dimensions"), **not** "upload failed". |

### Known fixture-environment caveat (NOT a code defect)

Re-loading a **just-committed gateway derivative** into the field crop modal
**in the same session** can yield an `<img>` stuck at `naturalWidth=0` even though:
the file is a valid `RIFF…WEBP` (verified), returns HTTP 200, `createImageBitmap`
decodes it to full dimensions, and a **fresh `Image()` loads the identical URL
fine** (`nat:[2048,1229]`). The pre-existing field **Crop** button on an asset
that predates the session loads perfectly (`perry.webp` → 450×291), and after a
rebuild the saved derivative renders fine (`perry-500x300.webp` → 500×300). So the
0×0 is a **local-serve cache/timing race** between the just-written file and the
modal's immediate `<img>` load, orthogonal to the gateway code. In a real
deployment the provider write and the served origin are the same store; if it ever
surfaces there, the fix is a cache-bust query on the crop modal's source — a
one-line change deferred out of this hardening slice (no product-behaviour change).

## A2. Slice 6 browser regression (post-review hardening)

Same environment (fork binary serving `crop-fixture`, local mode). Failure injection via a
FileReader patch (`FAIL-*` names error; `SLOW-*` delay 3s, with a per-name read counter) and a
one-shot `/postlocal` 500 fetch patch.

| Scenario | Result |
|---|---|
| Skip remaining spares approved work | approve image + auto-resolved PDF, "Skip remaining" on the next image → both stayed staged, Save ENABLED, committed; PDF byte-identical; skipped image never written |
| resolved + failed + queued → Skip remaining | failed row AND queued dropped; resolved item alone remained, savable |
| Failed item visible + recoverable | "⚠️ FAIL-doc.pdf — Failed to read blob" + Remove rendered persistently; Save disabled until Remove → then ENABLED with approved work intact |
| Mixed one-image batch affordance | single image + queued PDF still shows "Skip remaining" (batch-based skippable count, not images-only) |
| Single-image queue backdrop | inert ("queueMode" explicit — position display hidden, Skip-remaining hidden, modal stays) |
| Per-image modal reset | Crop ticked on image 1 (heading "Crop image"); image 2 opened RESET ("Optimise image", unchecked) |
| Session survives tab switch | Upload→Library→Upload mid-review: staged intact, awaiting image's modal re-bound |
| Cross-remount single read | SLOW pdf read spanning a tab switch → read counter = **1**; payload staged once (session-wide claim, not per-instance counter) |
| Teardown during processing | media modal closed mid-read → read completed but late result DROPPED (nothing staged, nothing on disk) |
| Save failure → retry | one-shot 500: `media[]` unpolluted, staged batch retained past the 900ms reset AND across a tab switch; retry committed both files |
| Field multi-drop | notice "Only one file can be used here — using <name>." shown; single file used; input `multiple=false` |
| Field regression | backdrop at rest cancels (crop modal only); one-click "Use optimised image" → eager commit → media modal closes → field crop modal opens (Slice 2 handoff intact) |
| File-input reset | `input.value === ''` immediately after selection (same-file re-selection fires) |
| Invariants | content JSON `data:image` = 0; 0 blob-backed images after close; no app console errors; no bijection-violation logs |

## B. Recorded live-provider smoke — 2026-07-02, `f3db27d`

These results were recorded against **disposable local server instances**, using
the actual `gitea.js` / `gitlab.js` provider modules from `feat/image-crop` at
**`f3db27d`**. Build-only imports were stubbed; HTTP requests were real. This was
agent-run provider smoke, distinct from owner-run UI testing. The original result
ledger and harnesses are retained in the GiteaPlenti tracking repository under
`docs/upstream/live-smoke/`; the dated results are summarized here so the PR's
citation is self-contained.

| Provider / environment | Recorded result |
| --- | --- |
| Gitea 1.26.4, native local server | 11/11 assertions passed: derivative upsert created on first upload and updated with SHA on re-upload; duplicate raw create threw; an early successful upsert remained after a later batch failure, with no `onSave` success signal |
| GitLab CE 19.1.1, local Docker amd64 under Rosetta | 10/10 assertions passed: derivative create then update via HEAD metadata; duplicate raw create rejected; a failing mixed batch persisted neither action and emitted no `onSave` success signal |

The Gitea run also observed that a contents `PUT` without a SHA created a missing
file on that server version. This is a dated server observation, not a guarantee
for every Gitea version or a substitute for explicit upsert resolution.

**Scope and limitations:** this evidence predates Slice 7 and its review fixes.
It does not establish a current-HEAD live rerun, current field/upload UI behavior,
or successful retry after partial persistence. Between `f3db27d` and the reviewed
`3401763`, the Gitea/GitLab module changes are upstream #375's optional `apiBaseUrl`
overrides; Groups 1–3 did not change their save logic. That source comparison does
not validate the new endpoint overrides against a live host. Current synthetic,
instrumented-browser, and ordinary-browser evidence is separated in the acceptance
matrix.

**Post-Draft correction (2026-09-13):** `838202d` subsequently changes Gitea's
ordering to include raw/as-is media `create` writes before content, rather than
prioritizing only derivative `upsert` items. Six new assertions across failure
and mixed-write scenarios exercise the actual provider with mocked HTTP responses.
They are current contract evidence, **not a new live-Gitea run**. The historical
table above remains tied to `f3db27d`; it does not establish the new ordering.
The later focused live run in §C now verifies that ordering at `60e1885`.

### B1. What partial persistence means for retry

Gitea makes sequential per-file commits. If an early raw `create` succeeds and a
later item fails, the existing path can reject the same `create` on retry. The
recorded tests establish duplicate-create rejection and partial persistence;
they do not claim an atomic rollback or automatic reconciliation. A failed batch
remaining available for review/retry is not a guarantee that retry succeeds
without addressing already-written files. GitLab's batch commit is atomic.

### B2. Checklist for a future live rerun

The unchecked items below are a **future two-provider rerun checklist**, not a
statement that the dated results above are missing. The newer focused Gitea
result is in §C; GitLab has not been rerun since the dated evidence above.

The provider **request contract** is unit-tested with mocked `fetch`
(`scripts/test-providers.mjs`): per-item create/upsert resolution, GitLab atomic
batch + HEAD-resolve, Gitea derivative-upsert-before-content ordering, and the failure
behaviour (metadata-probe failure aborts before writes; `onSave` never fires on
partial failure). Record the source SHA, date, server versions, and result when
these checks are next run against live hosts:

- [ ] **GitLab** — upload a derivative, then re-upload the *same* hashed name →
      the second commit resolves to `action: update` (HEAD found → `last_commit_id`
      sent), no 422. Atomic single-commit (content+media in one POST).
- [ ] **Gitea** — same derivative re-upload → GET finds the blob → `PUT` with the
      resolved `sha` (update), no 422. Media write precedes content write.
- [ ] **Passthrough duplicate** — re-upload a PDF/SVG/GIF of the same name → stays
      `action: create` (NOT silently overwritten just because the feature exists);
      an existing-file 422 is surfaced, not swallowed.
- [ ] **Gitea partial-batch caveat** — a mixed batch is **sequential**: if an early
      raw `create` (e.g. a PDF) commits and a later item fails, retrying the same
      `create` may now 422 because the file exists. This is a documented limitation,
      NOT atomic retry (see ADR 0001). Confirm the UI surfaces the partial state
      rather than claiming a clean rollback.

These provider checks need server credentials. The acceptance matrix labels the
other verification layers separately; an automated or local-fixture pass is not
a live-provider pass.

## C. Focused live Gitea run — 2026-09-13, `60e1885`

**26 assertions passed, 0 failed**, using the actual Gitea provider at
`60e18859837ff857b0147b7b928cd6d8f3ac9e4d` against a fresh **Gitea 1.26.4**
native localhost server with isolated SQLite data and a new private repository.
The provider ran as a verified **non-admin** user with `read:user,write:repository`.
A separate same-user setup token with `write:user` created the repository and was
never passed to the provider under test.

Build/generated imports and the user-store/history adapters were supplied by the
harness. Every HTTP request reached the real server; the observer recorded only
method/path/status. No response stubs, simulated faults, or browser UI were used.
Failures below were real duplicate-create rejections from Gitea.

| Scenario | Observed live result |
| --- | --- |
| Mixed raw PDF, derivative upsert, as-is WebP, and existing content | Media POSTs returned 201 in their input order, then content PUT returned 200; byte-identical media and correct JSON verified through the contents API; one success signal |
| Re-derive existing derivative | GET resolved SHA, PUT returned 200, stored bytes and SHA changed; one success signal |
| Duplicate raw PDF before existing page update | Media POST returned 422; no content write, existing content SHA/bytes unchanged, no success signal |
| Duplicate as-is WebP before new page create | Media POST returned 422; no content write, new page remained absent, no success signal |
| Later failure after an early raw create | Early POST returned 201 and remained persisted; later duplicate returned 422; referencing content remained absent, no success signal |
| Naive retry of that partial batch | First POST now returned 422 because the early file existed; content remained absent, no success signal |
| Slash-prefixed nested media path | Media and content POSTs returned 201 in the correct order; canonical media bytes verified |

Provider source SHA-256:
`786ec2a49e9c49d7335a2e9ae7ef84a661eed625d99358c37176144637e43c61`.
The reusable harness and redacted method/path/status ledger are retained in
GiteaPlenti at `docs/upstream/live-smoke/live-smoke-gitea-current.mjs` and
`docs/upstream/live-smoke/2026-09-13/gitea-results.json`.

This upgrades the new ordering from mocked-HTTP contract proof to **live-server
provider proof at the named commit**. It does not prove browser login/upload UI,
canvas behavior, custom API-base overrides, GitLab at current HEAD, or atomic
Gitea rollback. The partial-write/retry limitation was confirmed, not removed.
The normal five suites remain 175 checks; these 26 live assertions are reported
separately rather than folded into that count.
