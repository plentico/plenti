<script context="module">
  export const sortByDate = (items, order) => {
    items.sort((a, b) => { 
      // Must have a field specifically named "date" to work.
      // Feel free to extend to other custom named date fields.
      if (a.fields.hasOwnProperty("date") && b.fields.hasOwnProperty("date")) {
        let aDate = new Date(a.fields.date);
        let bDate = new Date(b.fields.date);
        if (order == "oldest") {
            return aDate - bDate;
        }
        return bDate - aDate;
      }
    });
    return items;
  };
</script>