<script>
    export let field, label;

    let isDate = date => (new Date(date) !== "Invalid Date") && !isNaN(new Date(date));
    //let makeDate = date => new Date(date).toISOString().split('T')[0];
    //let ogfield;
    let makeDate = date => {
        //ogfield = field;
        //field = new Date(date).toISOString().split('T')[0];
        //console.log(field);
        //return field;
        return new Date(date).toISOString().split('T')[0];
    }
    let bindDate = date => {
        console.log(date);
        //field = makeDate(date);
        //let y = new Date(ogfield);
        //let x = new Intl.DateTimeFormat('en-US');
        //field = x.format(y);
        //field = date.format(date);
        field = formatDate(new Date(date), 'mm/dd/yy');
    }
    function formatDate(date, format) {
        const map = {
            mm: date.getMonth() + 1,
            dd: date.getDate(),
            yy: date.getFullYear().toString().slice(-2),
            yyyy: date.getFullYear()
        }
        return format.replace(/mm|dd|yy|yyy/gi, matched => map[matched])
    }
</script>

{#if field === null}
    <div>{field} is null</div>
{:else if field === undefined}
    <div>{field} is undefined</div>
{:else if field.constructor === "".constructor}
    {#if isDate(field)}
        <!-- {makeDate(field) || ""}-->
        <!-- <input type="date" bind:value={field} on:input={resetDate} /> -->
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