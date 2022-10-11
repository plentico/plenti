<script>
    import { isTime, inputFormatTime, displayFormatTime } from '../date_checker.js';

    export let field;

    const bindTime = time => field = displayFormatTime(time, field);    
</script>

<input
    type="time"
    value={isTime(field) ? inputFormatTime(field) : null}
    on:input={time => bindTime(time.target.value)}
    step="{field.split(':').length > 2 ? 1 : 60}"
    required
/>