<script>
    import { isTime, makeTime, formatTime } from '../date_checker.js';

    export let field, schema, parentKeys;

    const bindTime = time => {
        field = formatTime(time, field);
    }
    
    let showSeconds;
    if (schema[parentKeys].hasOwnProperty("options")) {
        showSeconds = schema[parentKeys].options.includes("seconds");
    }
</script>

<input 
    type="time"
    value={isTime(field) ? makeTime(field) : null}
    on:input={time => bindTime(time.target.value)}
    step="{showSeconds ? 1 : 60}"
    required
/>