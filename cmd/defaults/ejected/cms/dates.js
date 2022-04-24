/**
 * 6/7/2008
 * 6/7/08
 * 6-7-2008
 * 6-7-08
 * 6.7.2008
 * 6.7.08
 */
const reDateNumeric = new RegExp("^(0?[1-9]|1[0-2])(/|-|\.)(0?[1-9]|1[0-9]|2[0-9]|3[0-1])(/|-|\.)(([0-9][0-9])?[0-9][0-9])$");

/**
 * Saturday, June 7, 2008
 * Saturday, Jun 7, 2008
 * Sat, June 7, 2008
 * Sat, Jun 7, 2008
 * 
 * (with or without commas and case insensitive)
 */
const reDateWritten = new RegExp("^(\\b((Monday|Mon)|(Tuesday|Tue)|(Wednesday|Wed)|(Thursday|Thu)|(Friday|Fri)|(Saturday|Sat)|(Sunday|Sun))\\b(,?) )?\\b((January|Jan)|(February|Feb)|(March|Mar)|(April|Apr)|May|(June|Jun)|(July|Jul)|(August|Aug)|(September|Sep)|(October|Oct)|(November|Nov)|(December|Dec))\\b (0?[1-9]|1[0-9]|2[0-9]|3[0-1])(,?) ([0-9][0-9][0-9][0-9])$", "i");

export const isDate = date => (new Date(date) !== "Invalid Date") && !isNaN(new Date(date)) && (reDateNumeric.test(date) || reDateWritten.test(date));
export const makeDate = date => new Date(date).toISOString().split('T')[0];
export const formatDate = (date, format) => {
    let parts = date.split('-');
    let year = parts[0];
    let month = parts[1];
    let day = parts[2];

    let days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
    let daysFull = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];
    let months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    let monthsFull = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

    if (reDateNumeric.test(format)) {
        let replacements = format.match(reDateNumeric);
        let delimeter = replacements[2];
        let yearFull = replacements[6];
        year = yearFull !== undefined ? year : year.slice(-2);
        return Number(month) + delimeter + Number(day) + delimeter + year;
    }
    if (reDateWritten.test(format)) {
        let replacements = format.match(reDateWritten);
        let delimeterDayOfWeek = replacements[10] !== undefined ? replacements[10] : '';
        let dayPos = new Date(date).getDay();
        let dayOfWeek = replacements[2] === undefined ? '' : replacements[2].length > 3 ? daysFull[dayPos] : days[dayPos];
        let space = dayOfWeek ? ' ' : '';
        let monthName = replacements[11] === undefined ? '' : replacements[11].length > 3 ? monthsFull[month -1] : months[month -1];
        let delimeterDay = replacements[24] !== undefined ? replacements[24] : '';
        return dayOfWeek + delimeterDayOfWeek + space + monthName + ' ' + Number(day) + delimeterDay + ' ' + year;
    }
    // Can't find format
    return date;
}