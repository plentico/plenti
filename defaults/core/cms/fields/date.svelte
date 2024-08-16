<script>
    import { isDate, makeDate, formatDate } from '../date_checker.js';

    export let schema, parentKeys, field;

    let today;
    if (schema && schema[parentKeys]?.options) {
        today = schema[parentKeys].options.includes("today");
    }

    if (today) {
        let date = new Date();
        date = makeDate(date);
        field = formatDate(date, field);
    }

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
