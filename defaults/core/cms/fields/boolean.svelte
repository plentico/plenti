<script>
    export let field, label;
</script>

<label><input id="{label}" type="checkbox" bind:checked={field} /> {field}</label>