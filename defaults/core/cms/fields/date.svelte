<script>
    import { isDate, makeDate, formatDate } from '../date_checker.js';

    export let schema, parentKeys, field;

    let today, todayPlusN, todayMinusN, N;
    if (schema && schema[parentKeys]?.options) {
        today = schema[parentKeys].options.includes("today");

        const regex = /^today\s?(\+|\-)\s?([0-9]+)$/g;
        const match = schema[parentKeys].options.find(option => regex.test(option));
        if (match) {
            const [, operation, daysOffset] = new RegExp(regex).exec(match);
            todayPlusN = operation === "+";
            todayMinusN = operation === "-";
            N = parseInt(daysOffset);
        }
    }

    if (today) {
        let date = new Date();
        date = makeDate(date);
        field = formatDate(date, field);
    }
    if (todayPlusN) {
        let date = new Date();
        date.setDate(date.getDate() + N);
        date = makeDate(date);
        field = formatDate(date, field);
    }
    if (todayMinusN) {
        let date = new Date();
        date.setDate(date.getDate() - N);
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
