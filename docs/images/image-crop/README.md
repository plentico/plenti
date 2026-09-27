# Image-crop field controls

Updated on 2026-09-13 with the owner-approved **Edit crop** label, serving a
disposable copy of `crop-fixture`. The fixture uses Perry for both sample fields;
the crop field has a 500×300 WebP schema, while the other field has `crop:false`.
These are actual browser hover states, without CSS overrides or image generation.

- `edit-crop-hover.png`: Change Media and Edit crop on the configured field.
- `change-only-hover.png`: Change Media alone on the field without crop enabled.

Edit crop is an action label, not an indicator of saved crop history. Browser
checks confirmed the same label before cropping, after applying/saving, and after
reload. The former session-dependent Crop Image / Re-crop Image wording is gone.

The screenshots document the implemented UI. The processing/deferral behavior is
described in #364; the Crop-only UI is the proposal in the author's follow-up,
without a separate maintainer confirmation visible in the issue thread at capture.
