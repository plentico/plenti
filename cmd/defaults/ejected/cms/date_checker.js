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

/**
 * 14:30
 * 02:30pm || 02:30 pm || 02:30PM || 02:30 PM
 * 2:30pm || 2:30 pm || 2:30PM || 2:30 PM
 * 14:30:45
 * 02:30:45pm || 02:30:45 pm || 02:30:45PM || 02:30:45 PM
 * 2:30:45pm || 2:30:45 pm || 2:30:45PM || 2:30:45 PM
 */
const reTime = new RegExp("^(0?[1-9]|1[0-9]|2[0-4])(:)([0-5][0-9])((:)([0-5][0-9]))?((\\s)?(pm|PM|am|AM))?$");

export const isTime = time => reTime.test(time);
export const inputFormatTime = time => {
    let replacements = time.match(reTime);
    let hour = replacements[1] === undefined ? '' : replacements[1];
    let minute = replacements[3] === undefined ? '' : ':' + replacements[3];
    let second = replacements[6] === undefined ? '' : ':' + replacements[6];
    // Get am / pm if set in initial string
    let period = replacements[9] === undefined ? '' : replacements[9];
    // Check if PM is used
    if (period === "pm" || period === "PM") {
        // HTML input needs 24 hour format
        hour = Number(hour) === 12 ? 12 : Number(hour) + 12;
    }
    // Check if AM is used
    if (period === "am" || period === "AM") {
        // Military format starts at 00 instead of 12am 
        if (Number(hour) === 12) {
            hour = "00";
        }
        // Time input needs leading zero
        if (hour.length !== 2) {
            hour = "0" + hour;
        }
    }
    // Remove am / pm so HTML input can read value
    return hour + minute + second;
}
export const displayFormatTime = (time, format) => {
    // Get parts from formatted HTML time input (the changing value from editor)
    let parts = time.split(':');
    let hour = parts[0];
    let minute = ':' + parts[1];
    let second = parts[2] === undefined ? '' : ':' + parts[2];

    // Recall the format used in original string from JSON
    let replacements = format.match(reTime);
    // Get the optional space before am or pm
    let space = replacements[8] === undefined ? '' : ' ';
    // Get AM or PM if available
    let period = replacements[9] === undefined ? '' : replacements[9];
    
    // Check if AM or PM was initially used
    if (period !== '') {
        if (Number(hour) >= 12) {
            // Convert back from military time to 12 hour format
            hour = Number(hour) === 12 ? 12 : Number(hour) - 12;
            period = period === period.toUpperCase() ? "PM" : "pm";
        } else {
            // Convert to number for comparison and removing leading zeros
            hour = Number(hour);
            if (hour === 0) {
                hour = 12;
            }
            period = period === period.toUpperCase() ? "AM" : "am";
        }
    }
    return hour + minute + second + space + period;
}