# ADR 0001 — Client-side image crop / scale / convert for Plenti

- **Status:** Proposed
- **Date:** 2026-06-30
- **Author:** Ben Waller
- **Upstream issue:** [plentico/plenti#364 — "Crop and scale images"](https://github.com/plentico/plenti/issues/364) (opened by @jimafisk, 2025-05-30)
- **Related work in flight:** [plentico/plenti#375 — CMS auth/token/endpoint overrides](https://github.com/plentico/plenti/pull/375) (from `plenti-core`, awaiting review — see "Interaction with #375")
- **Prototype reference:** the `Plenti-image-crop` site (`CROP_INTEGRATION_CHANGELOG.md`, `IMAGECROP_IMPLEMENTATION_PLAN.md`) — the original ejected prototype this feature ports from

---

## Context

### Origin of the prototype
A working image-crop prototype was built into the `Plenti-image-crop` **site** by ejecting Plenti's core CMS files and inlining the crop logic. It was authored against **Plenti v0.7.14** (Oct 30 2024). The prototype proved the concept but:

- inlined ~2,800 lines into two core files — `core/cms/fields/media.svelte` (71 → 1,346) and `core/cms/file_upload.svelte` (205 → 1,727) — because **Plenti cannot import *new* `.svelte` files into an ejected site** (only files that shadow an existing default are compiled);
- carries ~96 `console.log` debug statements;
- has a known unresolved binding bug for object-format media nested through `Fieldset`;
- **outputs at the on-screen crop-box size (~400 px), not schema-defined dimensions**, and **preserves the source format** rather than converting — i.e. the two things #364 actually wants (`scale`, `convert`) are the gaps.

### Why it broke (the trigger for this work)
Homebrew bumped the local `plenti` binary from v0.7.14 → **v0.7.21**. Two upstream changes broke the ejected site:
1. **v0.7.15** refactored publishing: `publish.js` + `post_local.js` → `core/cms/providers/{gitlab,gitea,local}.js`. The ejected `button.svelte` still imported the old paths → Gopack build failure.
2. A **`/postlocal` `file-path` validator was added after v0.7.14** (`cmd/serve.go`), restricting writes to `content/**.json`. The crop feature writes `media/*.webp` → rejected at runtime with a plain-text Go validator error (`Key: 'localChange.File'…`), which the client mis-parses as JSON.

> **Discovered during review:** because the existing media **upload** path also writes to `media/` through `Button → provider → /postlocal`, that validator has been silently rejecting **all local media uploads** since it was added, not just crop. Mirroring `media_checker.js`'s media set in the validator (D6) therefore *restores local uploads* in addition to enabling crop.

### Version boundary (established by inspecting upstream tags)
| Version | `publish.js`/`post_local.js` | `providers/` | `/postlocal` `file-path` validator |
|---|---|---|---|
| v0.7.14 (fork base) | ✅ present | ❌ | ❌ none (media writes allowed) |
| v0.7.15 | ❌ removed | ✅ | — |
| v0.7.21 (was installed) / v0.7.25 (latest) | ❌ | ✅ | ✅ `content/**.json` only |

`media.svelte` / `file_upload.svelte` / `dynamic_form_input.svelte` are **byte-identical across v0.7.14 → v0.7.25** — the media components never moved, so the work layers cleanly onto current Plenti.

### Maintainer intent
#364 (by Jim) asks for **exactly this**: client-side `<canvas>` crop/scale/convert, configured in `_schema.json`, written to Git before placement. His primary motivation is **performance** (a 5,000 px upload served where 500 px is needed). His sketched schema:
```json
{ "my_image_field": { "type": "media",
  "options": [ { "width": 500, "height": 300, "scale": true, "crop": true, "convert": "webp" } ] } }
```

### Future-architecture constraint (Pico/Pattr)
Plenti is migrating off Svelte to a native Go templating stack:
- **Pico** — Go compiler; `.pico` → `markup`/`script`/`style`; fence JS runs server-side via Goja. **No bundler, no ESM module resolution.** Client JS ships as plain static `<script>` assets.
- **Pattr** — Alpine-like reactive runtime (`pattr.js`); reactivity via `p-*` attributes / inline `onclick="{expr}"`.
- The new CMS (`pico-tests/site/views/cms.pico`) is a **generic key/value form generator with no media/upload/crop surface yet**, and **no save/commit mechanism yet**.
- **Cropping is client-side (Canvas); the migration is server-side rendering — orthogonal.** The transformation module never touches Pico's Go renderer.

---

## Decisions

**D1 — Do not pin/downgrade Plenti.** `brew pin` is global, would freeze every project and block `brew upgrade`. Target current Plenti instead. *(The local site was migrated to build on v0.7.21 — done.)*

**D2 — Build the feature in the `plenti-core` fork, not as ejected site files.** Only the source repo's `defaults/` allows adding *new* files (killing the inline-2,800-line constraint) and only it can carry the Go change. Branch `feat/image-crop` off `master`, isolated from `feat/cms-endpoint-overrides` (#375).

**D3 — Architecture = a browser *transformation module* + a thin host shell.** `crop-engine.js` is dependency-free and framework-agnostic, but it is **not "pure"**: it depends on the browser Canvas/DOM API and attaches one guarded `window` global. It **transforms only** — crop/scale/convert + filename — and **does not persist**. Validated against Pico/Pattr: the module survives the Svelte→Pico migration untouched; only the shell is rewritten.

**D4 — Delivery shape:** single dependency-free file, **ESM named exports + a non-clobbering `window.PlentiImage` attach** (`Object.assign(window.PlentiImage ||= {}, …)`), loadable as `<script type="module">`.
- Today (Svelte CMS): `import { transformImage, parseImageOptions } from './crop-engine.js'`.
- Future (Pico CMS): copy to `static/`, `<script type="module" src="/crop-engine.js">` beside `pattr.js`; inline Pattr handlers reach `PlentiImage.*`.

**D5 — Persistence is owned by the host, through Plenti's existing provider pipeline.** The module returns an encoded result (`{ blob, filePath, … }`); the **shell** converts the blob into the existing commit-item shape (`{ file, contents }`) and submits it through the provider-aware `Button → { postLocal | commitGitlab | commitGitea }`. This reuses Plenti's authentication, save-status UI, and per-provider commit logic, and works on local **and** GitLab/Gitea deployments — which a module-owned `/postlocal` call never would.
- The local provider (`providers/local.js`) currently parses failed `/postlocal` responses as JSON though the server returns plain text — a real existing bug; fix it to read `.text()` on `!res.ok`.
- **Eager vs deferred write is a Part-3 decision.** Default: **deferred** into the normal page-save commit (preserves provider behaviour, avoids orphaned cropped files when an editor cancels). Eager write (immediate media-library availability) is defensible but must call a **shared provider-dispatch helper**, not reproduce `Button` logic.

**D6 — The server change is required, and is *local-provider-only*.** GitLab/Gitea commit media through their Git APIs and never hit `/postlocal`; only the local dev provider does. So the `cmd/serve.go` change affects local serving only. Replace the single regex (which also wrongly matched `mediaevil.webp` / `contentbackup.json` — no required slash) with explicit validation that mirrors `media_checker.js`'s media set:
- `content/…\.json`, **or** `media/…\.<ext>` where `<ext>` ∈ the `media_checker.js` image+doc set (`jpg jpeg png webp gif svg avif apng pdf msword`). Mirroring `media_checker` keeps server and client agreeing on "what is media" and restores local uploads. *(Keep the two lists in sync — flag as a maintenance note in the PR.)* This **restores the existing upload policy** rather than asserting every type is safe: SVG can carry script and is served as-is, so sanitising user-uploaded SVG is a separate, out-of-scope concern flagged for the maintainer.
- Reject empty/absolute paths and any path that changes under `filepath.Clean` (blocks `..`).
- **Handler hardening (same commit, hunks splittable for Jim):** return non-2xx on JSON-decode / base64-decode / write failures (today it logs and still returns 200, so a client sees success after a failed write); `http.MaxBytesReader` with a named `maxPostLocalRequestBytes = 64 << 20` (base64 inflates a ~30 MiB image to ~40 MiB + JSON overhead — the limit is on the request, not the decoded image); `405` for non-POST; create the validator **once** (not per request); write files `0644` (not `os.ModePerm`); all errors via plain-text `http.Error` so `providers/local.js`'s `.text()` reader gets a clean message.
- **Delete addendum (found in owner manual review, 2026-07-12):** the upstream struct tags `Contents` as `required`, but a delete legitimately carries no contents (the client sends only `action`/`encoding`/`file`) — so every local media **delete** 400'd, in upstream as well as in the first D6 pass, which carried the tag forward. Fixed with `required_unless=Action delete`; contents remain mandatory for create/update (verified fail-closed).

**D7 — Scope = crop + scale + convert (the #364 core),** governed by the crop×scale matrix below. Output filename encodes **output dimensions + actual output extension**: `media/perry-500x300.webp` (re-crops strip a prior `-WxH` / `-cropped-WxH` suffix instead of stacking).

**D8 — Schema shape is the one open maintainer call.** `parseImageOptions(schema, fieldKey)` accepts **both** Jim's `options:[{…}]` and the prototype's `crop:{…}`, so the demo runs today and the final shape is decoupled from the module core.

**D9 — Output formats & browser fallback.** Supported **output** encoders: `jpeg`, `png`, `webp`, `avif`. **WebP/AVIF encoding is browser-dependent**; PNG is the guaranteed canvas fallback, and the **actual `blob.type` is authoritative** for the saved filename (so a webp request that silently falls back to PNG is saved—and named—as PNG, and the UI is told). **GIF is not an output target** (canvas cannot encode it). **Animated sources (gif/apng) flatten to a single still frame** when transformed — documented behaviour. (The server media-write allowlist in D6 is broader than this output set because it must also accept existing *uploads*.)

---

## Crop × scale semantics (the behaviour contract)

| `crop` | `scale` | behaviour |
|---|---|---|
| true | true | crop to the selection, then resize to configured `width`×`height`. **Reject** if the selection aspect ≠ output aspect (small tolerance). |
| true | false | crop to the selection; output = the selection's source-pixel dimensions (no resize). |
| false | true | whole image, **contain** within `width`×`height` — preserve aspect, **no padding, no upscale**. |
| false | false | whole image at original dimensions; format conversion only. |

**Edge cases:**
- `width` only → derive `height`; `height` only → derive `width`; **neither** → original dimensions; **both** → fit within the bounding box.
- `crop:true` **requires** a selection — error if absent; never silently fall back to the whole image.
- `crop:true` + `scale:true` with **only one** output dimension → derive the other from the selection ratio (no mismatch rejection needed).
- **contain** (`crop:false`+`scale:true`) treats `width`/`height` as **maximum** bounds, not a mandatory canvas: `5000×3000` into `500×500` → `500×300`; never letterboxes, never upscales beyond the source.
- Converting a transparent source (e.g. PNG) to JPEG fills a configurable `background` (default `#ffffff`) before drawing, since JPEG has no alpha.
- Output dimensions are always positive integers; an over-large request (`> MAX_OUTPUT_PIXELS`) is rejected.

---

## Architecture

```
crop-engine.js  (dependency-free browser transformation module — the durable asset; D3)
  parseImageOptions(schema, fieldKey)         -> normalized opts | null            // adapter (D8)
  renderImage(img, selection, output)         -> { blob, mime, width, height }     // crop+scale+convert, 1 drawImage
  outputFilename(srcPath, {width,height,ext}) -> "media/perry-500x300.webp"
  transformImage(img, selection, opts, src)   -> { blob, filePath, width, height,  // applies the matrix; NO save (D5)
                                                   requestedMime, actualMime,
                                                   formatFallback, converted, bytes }

host shell (thin, per-framework)
  - now:   Svelte modal in defaults/core/cms/fields/ — crop UI, calls transformImage,
           builds { file, contents } and submits via Button -> provider (D5)
  - later: a .pico component + Pattr attributes + global-attach glue, same module

server cmd/serve.go (local provider only; D6)
  - explicit path validation mirroring media_checker.js + handler hardening
```

---

## Open questions (for Jim)

1. **Schema shape** — `options:[{width,height,scale,crop,convert}]` (his sketch) vs `crop:{…}` (prototype)? The module accepts both today; he picks the canonical one. *(Write-allowlist and no-stretch are resolved above: mirror `media_checker`; reject distortion.)*
2. **Field-schema role now that the Library optimises on ingest (the D12 Stage-2 gate)** — with optimisation happening as images enter the Media Library, should field schema options **continue to auto-process** a selected image (D10 + the `options[]` sketch), or become **placement-crop requirements only** (selecting media assigns the library asset unchanged; the field exposes an explicit **Crop** action)? Consequence A: field selection may optimise/crop again automatically. Consequence B: field selection assigns the path unchanged; the Crop button makes a placement-specific derivative. This is a public schema-contract change, so D12 **Stage 2 is held** until this is confirmed; Stage 1 preserves D10 unchanged.

---

## Required tests before the PR
1. `null` and `undefined` selection both use the whole image (regression: the `num(null)`→`0`→1×1 bug).
2. A selection at/beyond an image edge never yields zero width/height.
3. `scale:false` preserves the selected source-pixel dimensions.
4. `crop:false` ignores any supplied selection (whole-image contain).
5. `crop:true` with no selection is an error (no silent whole-image fallback).
6. Fractional schema dimensions become consistent integers (metadata == bitmap == filename).
7. Aspect mismatch (`crop:true`+`scale:true`, both dims) is rejected, not stretched.
8. Unsupported output encoding (e.g. webp on old Safari) yields a correctly **named PNG** + `formatFallback:true`.
9. `.jpeg → image/jpeg` is **not** reported as `converted`.
10. Transparent input → JPEG gets the configured background.
11. Path validation accepts `media/foo.webp`, rejects `mediaevil.webp`, `../media/foo.webp`, absolute paths, unsupported extensions.
12. A local write/decode error reaches the UI as a **failed** save (server non-2xx + provider `.text()`).
13. The transformed commit item saves through local **and** (smoke) the GitLab/Gitea provider dispatch.

---

**D10 — Auto-process on field update (the #364 behaviour).** The primary trigger is **selecting/uploading a new image into a schema-configured media field**, not a manual button:
- no image options → assign the path (unchanged);
- `crop:true` → auto-open the crop modal on the new source (cancel reverts to the previous value — an uncropped large image must not be left in a crop:true field);
- `crop:false` → auto-optimise immediately (contain/convert), no modal;
- a failed transform keeps the previous field value + shows an inline error.

This holds for **both entry points into a configured field**:
- *Library pick* (an existing path) — handled via the `changingMedia` reactive.
- *Fresh upload* (a `File`) — `file_upload.svelte` runs in a field-scoped mode (registered through the `fieldUploadHandler` store) that hands the `File` to the field **instead of eager-saving the original**; the field processes it and queues only the derivative, so the untouched original never reaches Git. The handler is cleared on modal close, so the standalone Media library and no-option fields keep their eager upload behaviour.

The manual Crop/Optimise button remains as a re-process surface. **Direct Media-Library crop (cropping an arbitrary library image with no field/schema context) is a separate, secondary enhancement** — it can't know which field's dimensions/format to enforce.

## Implementation plan (all committed on `feat/image-crop`)
- **Server (`cmd/serve.go`):** explicit `/postlocal` validation (mirrors `media_checker.js`) + handler hardening (D6).
- **`crop-engine.js`:** transform-only module + the matrix + dependency-free tests (D3/D7/D9).
- **Deferred-persistence foundation:** `providers/commit.js` (shared dispatch), per-item action/encoding in all providers, `pending_media.js` store, Button `beforeSubmit` merge (D5).
- **Field integration:** `image_crop_modal.svelte` + `media.svelte` (schema resolution, source-path distinction, object dual-format, preview persistence) + `media_checker.js` guard.
- **Auto-process on update (D10)** — for library picks and fresh uploads (`field_upload.js` store + `file_upload.svelte` field-scoped mode).

### Acceptance — verified in a clean fixture site (browser + disk)
1. ✅ Select a 5000px JPEG into `crop:true` `hero_string` → modal **auto-opens** on that source; confirm → derivative + content committed in **one** provider commit. Byte proof: 5000px JPEG → 500×300 WebP = **98.8%** (388,409 → 4,862 b).
2. ✅ Select it into `crop:false` `banner_contain` → **auto 500×300 derivative, no modal** = **97.7%** (388,409 → 9,110 b, jpg preserved — no `convert`).
3. ✅ **Upload** a fresh 5000px JPEG into `crop:true` `hero_string` → crop modal opens **on the upload before any request**; into `crop:false` `banner_contain` → derivative, no modal. Page save writes **only** the derivatives; the original `fresh-upload.jpg` **never appears in `media/`**. Cancel makes no commit and keeps the previous value.
4. ✅ Object media: `src` updated, `alt` preserved; re-crop uses the **original** source + replaces (no duplicate); GIF flattens to a still webp; ordinary fields unchanged.
5. ✅ Engine unit tests (22) + `/postlocal` curl matrix all green.

---

## Interaction with #375
#375 (CMS auth/token/endpoint overrides) touches the remote-commit/auth area; this feature's server change is **local-provider-only**, so they are largely orthogonal. `feat/image-crop` branches off `master` (no #375), and rebases as upstream lands.

---

## Status of work so far
Committed on `feat/image-crop`: ADR + pipeline doc → server hardening → `crop-engine.js` + tests → `providers/local.js` fix → deferred-persistence foundation → field-crop modal integration → auto-process-on-update → field-scoped upload pre-processing → **remote-provider upsert + Gitea media-first ordering + provider contract tests (D11)**.

The select **and** upload paths into a configured field now process the image before it is saved to the repository, and the local, GitLab, and Gitea commit contracts are covered by `scripts/test-providers.mjs`.

**Remaining before a PR:**
- ✅ **Remote-provider gaps CLOSED (D11 below).** The two remote-only gaps a code-trace found (the local provider hid them by overwriting) are fixed and covered by committed fetch-level contract tests — see **D11**. A live remote smoke test remains desirable *confirmation*, not the primary proof.
- ⏳ Schema shape (`options[]` vs `crop{}`) — settle with Jim (the parser accepts both today).
- ⏳ Optional follow-up: direct Media-Library crop controls (D10).
- ✅ **JS test location/CI decided** — follow the existing `scripts/test-crop-engine.mjs` convention: dependency-free `.mjs` under `scripts/`, run with `node scripts/<name>.mjs` (no framework, no `package.json`). Provider tests added as `scripts/test-providers.mjs`. (How these wire into CI is for the maintainer; locally they are `node`-runnable.)

---

**D11 — Remote commit gaps closed: provider-neutral `upsert` + Gitea media-first ordering + commit-level success.** The mixed content+media commit model (D5) exposed correctness gaps on the *remote* providers (local hides them by overwriting). All are now fixed in `feat/image-crop`, verified by `scripts/test-providers.mjs` (22 fetch-level cases, passing under Node 20 and Deno):

- **Gap 1 — media `create` vs `update` on an existing path.** `pendingMedia.toCommitItems()` previously hardcoded `action:'create'`; re-deriving the *same* filename in a *later* session collided (GitLab rejected the atomic commit; Gitea POST 422). **Fix:** `toCommitItems()` now emits a provider-neutral **`action:'upsert'`** that each provider resolves against the live repo — it is never sent to a remote API verbatim:
  - **GitLab** — `HEAD` the file path → `404` = `create`, `200` = `update` (+ the `X-Gitlab-Last-Commit-Id` the batch Commit API requires for updates). All upserts resolve *before* the single commit, so it **stays one atomic request** (all-or-nothing). Any non-404 metadata status (auth/permission/server) **aborts** — never treated as "absent".
  - **Gitea** — `GET` the file → `200` = `update` (use the returned `sha`), `404` = `create`. Non-404 aborts. (Reuses the GET-for-sha the provider already did for content updates.)
  - **Local** — maps `upsert`→`create` on the wire (the `/postlocal` write overwrites; the server validator only knows create/update/delete).
- **Gap 2 — Gitea is non-atomic.** Its contents API is per-file, so a mixed save is N sequential commits. **Fix:** commit **media (the upserts) before content** — if a media write fails the loop aborts before the content (which references that derivative) is written. GitLab (single `actions[]` commit) and local are unaffected.
- **Gap 3 — Gitea `onSave` fired per file, not per commit.** Media-first ordering turned a pre-existing untidiness into a real bug: the post-success `onSave`/route-push sat *inside* the per-file loop, so a successful media write signalled "saved" to the UI **before** a possibly-failing content write — contradicting the retryable guarantee. **Fix:** `onSave`/`onDelete`/route-push moved to **commit-level** (after every file commits). Asserted by 4 cases: multi-item success fires `onSave` exactly once; a media failure and a content-after-media failure each fire it zero times; single-item behaviour unchanged.
- **Accepted residual limitation (documented):** on Gitea, if media succeeds but the *content* write then fails, an **orphan derivative** can remain in `media/`. This is clearly safer than the inverse (content pointing at a missing image) and matches Gitea's existing always-sequential commit behaviour — no new failure mode is introduced.
- **Deferred (separate provider-wide work, NOT this PR):** migrating the whole Gitea provider to its **multi-file commit API** (`/contents` batch with `operation:"upload"`) would make Gitea commits atomic like GitLab's. That benefits *every* Gitea commit, not just image-crop, and changes the provider's commit model + needs compatibility detection + a fallback — so it belongs in its own provider-focused issue/PR. Recorded here so it isn't lost.

---

**D12 — The Media Library is the image-optimisation *gateway* (Stage 1, additive; a schema-semantics change is Stage 2, pending Jim).**

Side-by-side testing of this fork against the original prototype showed the feature implemented **field-specific** processing (D10) but was missing the general ingestion gateway #364 actually describes: *upload into the Media Library → it is optimised at that point*. The Library is the funnel through which images enter the CMS; authors then *select* an already-optimised asset onto a field. This decision adds that gateway **without changing the field-schema contract** (that is held for Jim, below).

**The three-representation invariant (load-bearing).** An image exists in exactly three forms that are never interchanged:

```
Preview     blob:… / data:image/…;base64,…        in-memory, for <img> only
Transport   { action, encoding:'base64', file:"media/name-<hash>-WxH.webp",
              contents:"data:image/webp;base64,…" }   commit item; provider strips the prefix
Persisted   media/name-<hash>-WxH.webp             the string in the grid AND field content JSON
```

The grid and every field value hold the **persisted path**, never a transport data URL. A pre-existing latent bug pushed `item.contents` (a data URL) into `media[]`; it was masked because the grid's `isImage` accepts both a path and a data URL. `addUploadsToLibrary()` now enters **only paths** (`new Set` deduped) — one rule for optimised derivatives *and* passthrough.

**Stage 1 — what this PR builds (no schema-contract change):**
- **`crop-engine.js`** — `LIBRARY_OPTIMISE_DEFAULTS` (`{crop:false, scale:true, convert:'webp', quality:0.82, maxWidth:2048, maxHeight:2048}`) + a new **`maxWidth`/`maxHeight` "contain within a MAX edge"** branch (distinct from exact `width`/`height`; mixing the two is rejected) + `transformImage` now returns the normalised `sourceRect`. The exact-dimension field path (D7) is untouched.
- **Eager canonical save.** Library uploads commit immediately (reusing the shared `commit()` from D11), NOT the deferred `pendingMedia` flow (that stays the field-crop mechanism). Standalone queues derivatives for a "Save Media" batch; **field-launched saves in one click** and auto-returns the persisted path.
- **Collision-resistant hashed identity** — `media/name-<hash16>-WxH.webp` via a deterministic **versioned two-stage SHA-256**: `sourceDigest = SHA-256(bytes)`, then `SHA-256(canonicalJSON{version, sourceDigest, sourceRect, maxWidth, maxHeight, format, quality})`. Same source + same settings → same path (safe `upsert` update); different source **or** different crop → different path (no silent overwrite of a live asset). The **field**-derivative name (`outputFilename`, D7) is deliberately left un-hashed — hashing is only the Library-ingestion identity.
- **Per-item action** (D11) carries the mix: generated derivatives `upsert`, raw passthrough (PDF/SVG/GIF) `create`, in one commit list. Passthrough retains original bytes + extension (animation preserved).
- **Entry-point ownership** — an explicit `uploadContext` (`{kind:'standalone'}` vs `{kind:'field', onSavedPath}`), **never inferred** from `changingMedia`. On a field-launched success, `file_upload.svelte` **emits** the path; `media_modal.svelte`/`admin_menu` owns the add-to-library + close + **context reset** (a child can't reset a prop it doesn't own — else a later standalone upload assigns to a stale field), then hands the path to the field, which runs its existing D10 schema processing.
- **Optional Crop toggle — standalone only.** A field with `crop:true` already opens its own placement crop *after* the gateway returns, so the gateway is optimise-only for field uploads (exactly one interactive crop total).
- **Standalone multi-file queue** — `upload_queue.js`, a DOM-free state machine (`queued → preparing → awaiting_decision → processing → resolved → saved`; terminal `failed`/`cancelled`). Deterministic order, one active item at a time, Save disabled while anything is unresolved, idempotent object-URL revocation, Cancel-current/Cancel-all with a real `AbortController`.
- **`button.svelte`** gains an additive `disabled` prop (default false) to gate "Save Media" while unresolved.

**The ownership boundary (asserted by the two forced-failure tests):** once the provider commit succeeds, the **upload has succeeded**. A subsequent field-schema failure is an **assignment/field-processing failure**, not an upload failure — the canonical asset stays in the library and the previous field value is retained. Conversely a provider failure *before* the canonical save changes nothing (field, `media[]`, disk) and reports an upload failure. Both proven in the browser fixture (forced 500 / forced transform failure) and, on the request side, by `scripts/test-providers.mjs` (abort → no write; `onSave` never fires on partial failure).

**Stage 2 — HELD, pending Jim's decision (do NOT build until confirmed):** whether field schema options should keep **auto-processing** a selected image (the D10 behaviour + Jim's original `options:[{width,height,scale,crop,convert}]`), or become **placement-crop requirements only** — selecting media assigns the library asset unchanged and the field exposes an explicit **Crop** action. This would redefine `crop:false`/`scale`/`convert` field semantics — a public schema-contract change, so it is isolated behind this gate. The raw-File→`uploadContext` handoff is already done in Stage 1 and is **not** part of Stage 2.

**Tests & evidence (all in `feat/image-crop`):** engine 27, gateway (`test-library-optimise.mjs`) 12, queue (`test-upload-queue.mjs`) 64, providers 22 — all green; `go build ./...` exit 0. Full traceability in [`docs/364-acceptance-matrix.md`](../364-acceptance-matrix.md); browser + remote-smoke evidence in [`docs/364-remote-smoke.md`](../364-remote-smoke.md).

**D12 addendum — Slice 6 corrections (post-review hardening).** A high-effort code review of the
first gateway implementation confirmed 15 findings, rooted in two unsynchronized ledgers (the
queue's lifecycle vs the staged commit list) plus edge-path cancellation/remount gaps. The
corrections, all verified in the browser matrix:

- **Session ownership**: the upload session (queue + a keyed payload store,
  queue-item id → transport) is owned by `media_modal.svelte` and SURVIVES Upload↔Library tab
  switches; `file_upload.svelte` is a remountable view driving it through owner-controlled
  operations (one reactive root — no per-mutation invalidation convention to forget).
- **Single lifecycle ledger with keyed staged payloads**: resolving an item and staging its
  transport is one atomic `complete()` (ownership check → stage → transition → rollback on
  refusal → rebuild); the Save list is DERIVED from resolved items through the store, and the
  Save gate fails closed if the bijection is ever breached.
- **`skipRemaining()`** (UI: "Skip remaining") replaces `cancelAll()`: retains ONLY resolved
  items — approved work stays savable — and drops pending AND failed items, so one unreadable
  file can no longer brick a batch. **Failed items render in a persistent status area with a
  Remove control** (the queue's `remove()` finally wired).
- **Session-wide run claims** (`claimRun` → `{ itemId, token }` on the queue itself): a later
  claim from any component instance invalidates an earlier chain, so a remount cannot double-
  drive an item (browser-proven: a delayed read spanning a tab switch performed exactly ONE
  read); teardown invalidates the claim so late completions are ignored — cancellation is
  logical invalidation, not physical abort.
- **Explicit teardown on modal close**: skip non-resolved → promote resolved payloads into the
  surviving (admin_menu-owned) commit list → destroy queue + store. The queue-null Save
  fallback can therefore only ever see approved payloads.
- **`queueMode` is an explicit modal prop**: the backdrop dismiss is inert in queue mode (an
  accidental click never skips a file) and inert while processing in field mode (a cancel can
  never race an in-flight commit — closes the orphaned-committed-file path). The
  Skip-remaining affordance derives from the count of actionable (skippable) items, never raw
  batch length. `{#key}` remounts the crop modal per image (the Crop toggle and pan/zoom reset).
- **`media[]` is mutated only after provider success** (Button's success-only `afterSubmit`),
  and the shared Button gains an OPT-IN `retainCommitListOnFailure` so a failed Save Media keeps
  the staged batch for retry (every other Button call site unchanged).
- **Accepted residual** (documented): an outer-modal close during a FIELD eager-commit cannot
  un-send the commit — a late success leaves an orphaned committed file with no path delivery
  (same class as the D11 Gitea orphan; window ≈ one commit round-trip).
- **Plenti pipeline constraint discovered**: `cmd/build/compile.go`'s regex import-rewriter
  corrupts a component whose compiled SSR contains MORE THAN ONE printer-wrapped import (the
  greedy multi-line branch spans from the first `import {` to the LAST line-start `} from`).
  Mitigation in this branch: keep hand-written import lines under ~80 chars (noted in the
  component). A proper Go-side fix (non-greedy/anchored matching) is upstream-worthy but out of
  scope here.

---

**D13 — Field model from the maintainer discussion (supersedes the D12 Stage-2
placeholder).** Jim answered the schema/architecture question on #364 (comment
thread, 2026-07-03). The processing/deferral model below follows that discussion.
The Crop-only controls are the author's follow-up proposal, adopted for this
implementation; no separate maintainer reply approving that UI detail is visible
in the public thread checked on 2026-09-13. This records the implementation
decision, not evidence of an additional public approval.

- **Selection into a schema-configured field auto-processes** (his choice — the D10 semantics
  Stage 1 preserved turn out to be the confirmed direction), with a NEW **conformance
  short-circuit**: an asset that already meets the field's spec is referenced directly, no
  derivative copy (`conformsToImageOptions` — conservative: `convert` requires a matching
  extension with jpg/jpeg as one format; `crop:true` needs BOTH dims and an exact match;
  underdetermined crops never short-circuit; `crop:false`+`scale` conforms within bounds;
  `scale:false` is format-only).
- **Field-launched uploads are DEFERRED** (his "consolidate media swapping to one commit"):
  the gateway still optimises clientside immediately, but the canonical asset is staged in
  `pendingMedia` and flushes WITH the page save — canonical + any placement derivative +
  content in one provider save operation (atomic on GitLab, sequential on Gitea);
  abandoning the edit before Save persists nothing. Deferred RAW passthrough keeps
  per-item `action:'create'` (Gitea/GitLab surface same-name conflicts at save; the existing
  local dev endpoint still overwrites, tracked as a separate backend fix).
  **Standalone Media-Library uploads keep the explicit eager "Save Media" batch**
  (no page-save moment exists there) — D12's eager-save scope narrows to exactly that surface.
- **Field UX (implemented proposal)**: automatic `scale`/`convert` is silent — deterministic, applied on
  selection, conformance-skipped — so there is **no Optimise button**. Crop-configured fields
  expose one explicit **Edit crop** action beside Change Media in a split hover overlay (the
  original prototype's visual, rebuilt); ordinary and `crop:false` fields show Change Media
  only. The owner-selected label "Edit crop" is consistent before and after a
  crop/save/reload; it does not infer persistent crop history from session metadata.
- **Ingestion conformance (owner-confirmed, 2026-07-12)** — the field-selection short-circuit's
  principle extended to the gateway itself: an upload already meeting the library defaults
  (target format + within the max edge) is added **as-is** — original bytes, original name (the
  raw-passthrough naming contract, not the hashed derivative identity), `create` action — never
  silently decoded/re-encoded, since a deliberately pre-optimised asset can come out larger.
  Standalone shows an explicit as-is review step (crop remains an opt-in that re-enters the
  derivative flow); field-launched conforming uploads skip the optimise modal and defer as-is.
- **Terminology reconciled** (his question): *optimise* = the no-interaction ingestion bundle
  (downscale-to-max `scale` + `convert` + quality); *convert* is one ingredient; *crop* is the
  one interactive operation.
- Side effect: the Slice-5 "just-committed WebP `<img>` race" caveat is structurally gone for
  the field flow — deferred assets always render from in-memory blobs until persisted.
- **Addendum (owner manual review, 2026-07-12) — the PAGE preview is now truthful too.** The
  first D13 pass only wired `pendingMedia` previews into the CMS field widget; the user's own
  page markup rendered the deferred PATH and showed a broken image until page save —
  contradicting the "clientside immediately (so the page preview is truthful)" commitment.
  `preview_patcher.js` (mounted at the CMS root, `admin_menu`) swaps any page `<img>`/`<embed>`
  whose src is a pending path to the entry's object URL and restores the original when the
  entry leaves the store; a MutationObserver covers page rerenders. Same review also made the
  field-error surface legible on dark trays (it rendered darkred-on-dark, so a failed selection
  looked like a silent no-op). A second review pass extended the same session-truthfulness
  principle to the Library: a page save now appends the just-persisted derivative paths to the
  in-session `media[]` (deduped, `mediaPrefix`-normalised) instead of waiting for a reload to
  pick up the regenerated media list — the patcher's blob covers the new grid tile while the
  site rebuild races the disk write.

Browser-proven on the fixture: field upload stages with ZERO commits; the page save issues
exactly ONE `/postlocal` request carrying `update:content + canonical + placement derivative`;
reload-before-save persists nothing; the conformance short-circuit assigned a 5000×3000 JPEG
directly into a `convert:'jpg', scale:false` field (no copy, no commit); standalone Save Media
unchanged (eager batch); no Optimise control anywhere; engine suite 39 (12 conformance cases).

**Review follow-up (2026-09-12, Groups 1 and 2).** D13's model is unchanged. Field
requests now lose ownership on a newer selection, picker reopen, or destruction;
upload probes also check the captured picker context. Guards cover both decode and
automatic field encoding. The page patcher records its last written URL separately
from the current canonical source, so a new selection replaces the old record and
both store removal and teardown restore the current source. PDF tiles keep a
separate canonical path for selection even when their `src` is a blob preview.
See the acceptance matrix's review-follow-up section for synthetic and instrumented
browser evidence. Library deletion resurrection was left for Group 3 below.

**Review follow-up (2026-09-12, Group 3).** A successful Library deletion retires
the matching committed pending-media entry and its preview URL before updating
the Library list. This prevents the retained entry from being appended again.
Unsaved replacements remain staged, and a later save of the same path can append
normally. This uses the existing deletion-success boundary rather than a permanent
append-once ledger; failures do not invoke retirement. The acceptance matrix
records synthetic tests and ordinary browser proof of save/delete/re-save on an
unchanged homepage filepath. The local create-collision backend defect is still
unfixed and disclosed, outside this contribution's scope.
