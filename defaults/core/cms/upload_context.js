// upload_context.js — the explicit Media-Library upload context.
//
// Who opened the picker, and (for a field) where to hand the canonical path back.
// onSavedPath is the existing handoff name; field uploads are staged until page save.
// Threaded media.svelte / admin_menu -> media_modal -> file_upload. NEVER inferred
// from changingMedia (a field with no image has changingMedia === '', which would
// read as standalone). The modal owner resets to STANDALONE on close so a later
// standalone upload can't fire a stale field callback.
//
//   { kind: 'standalone' }
//   { kind: 'field', onSavedPath(path) { … } }

export const STANDALONE_UPLOAD_CONTEXT = Object.freeze({ kind: 'standalone' });
