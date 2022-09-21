<script>
    import { makeDate, formatDate } from '../dates.js';

    export let field;

    const bindDate = date => {
        field = formatDate(date, field);
    }
</script>

<input type="date" value={makeDate(field)} on:input={date => bindDate(date.target.value)} />