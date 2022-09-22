<script>
    import { isDate, makeDate, formatDate } from '../dates.js';

    export let field;

    const bindDate = date => {
        field = formatDate(date, field);
    }
</script>

<input 
    type="date"
    value={isDate(field) ? makeDate(field) : null}
    on:input={date => bindDate(date.target.value)}
    required
/>