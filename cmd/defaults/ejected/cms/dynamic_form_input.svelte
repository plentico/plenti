<script>
    export let field, label;

    let isDate = date => (new Date(date) !== "Invalid Date") && !isNaN(new Date(date));
    let makeDate = date => new Date(date).toISOString().split('T')[0];
    let bindDate = date => {
        //console.log(date);
        //field = formatDate(new Date(date), 'mm/dd/yyyy');
        //field = date;
        field = formatDate(date, field);
    }
    let formatDate = (date, format) => {
        console.log('date = ' + date);
        let parts = date.split('-');
        let year = parts[0];
        let month = parts[1];
        let day = parts[2];

        let months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
        let monthsFull = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

        console.log('og date = ' + format);
        //let re = new RegExp("^(0[1-9]|1[0-2])\/(0[1-9]|1\d|2\d|3[01])\/(19|20)([0-9][0-9])");
        //let re = new RegExp("^(0[1-9]|1[0-2])/(0[1-9]|1[0-9]|2[0-9]|3[0-1])/(19|20)([0-9][0-9])");
        let re = new RegExp("^(0?[1-9]|1[0-2])/(0?[1-9]|1[0-9]|2[0-9]|3[0-1])/(19|20)([0-9][0-9])");
        if (re.test(format)) {
            return month + '/' + day + '/' + year;
        }
        return date;
    }
    /*
    function formatDate(date, format) {
        const map = {
            mm: date.getMonth() + 1,
            dd: date.getDate() + 1,
            yy: date.getFullYear().toString().slice(-2),
            yyyy: date.getFullYear()
        }
        return format.replace(/mm|dd|yy|yyy/gi, matched => map[matched])
    }*/
</script>

{#if field === null}
    <div>{field} is null</div>
{:else if field === undefined}
    <div>{field} is undefined</div>
{:else if field.constructor === "".constructor}
    {#if isDate(field)}
        <input type="date" value={makeDate(field)} on:input={date => bindDate(date.target.value)} />
    {:else if field.length < 50}
        <input id="{label}" type="text" bind:value={field} />
    {:else}
        <textarea id="{label}" rows="5" bind:value={field}></textarea>
    {/if}
{:else if field.constructor === true.constructor}
    <input id="{label}" type="checkbox" bind:checked={field} /><span>{field}</span>
{:else if field.constructor === [].constructor}
    <fieldset>
        <legend>{label}</legend>
        {#each field as value, key}
            <svelte:self bind:field={field[key]} {label} />
        {/each}
    </fieldset>
{:else if field.constructor === ({}).constructor}
    {#each Object.entries(field) as [key, value]}
        <div>
            <label>{key}</label>
            <svelte:self bind:field={field[key]} {label} />
        </div>
    {/each}
{/if}

<style>
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
    }
</style>