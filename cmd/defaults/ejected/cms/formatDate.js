export const formatDate = (date, format) => {
    let parts = date.split('-');
    let year = parts[0];
    let month = parts[1];
    let day = parts[2];

    let days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
    let daysFull = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];
    let months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    let monthsFull = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

    // 6/7/2008
    let re = new RegExp("^([1-9]|1[0-2])/([1-9]|1[0-9]|2[0-9]|3[0-1])/([0-9][0-9][0-9][0-9])");
    if (re.test(format)) {
        return Number(month) + '/' + Number(day) + '/' + year;
    }
    // 6/7/08
    re = new RegExp("^([1-9]|1[0-2])/([1-9]|1[0-9]|2[0-9]|3[0-1])/([0-9][0-9])");
    if (re.test(format)) {
        return Number(month) + '/' + Number(day) + '/' + year.slice(-2);
    }
    // 6-7-2008
    re = new RegExp("^([1-9]|1[0-2])-([1-9]|1[0-9]|2[0-9]|3[0-1])-([0-9][0-9][0-9][0-9])");
    if (re.test(format)) {
        return Number(month) + '-' + Number(day) + '-' + year;
    }
    // 6-7-08
    re = new RegExp("^([1-9]|1[0-2])-([1-9]|1[0-9]|2[0-9]|3[0-1])-([0-9][0-9])");
    if (re.test(format)) {
        return Number(month) + '-' + Number(day) + '-' + year.slice(-2);
    }
    // 06/07/2008
    re = new RegExp("^(0[1-9]|1[0-2])/(0[1-9]|1[0-9]|2[0-9]|3[0-1])/([0-9][0-9][0-9][0-9])");
    if (re.test(format)) {
        return month + '/' + day + '/' + year;
    }
    // 06/07/08
    re = new RegExp("^(0[1-9]|1[0-2])/(0[1-9]|1[0-9]|2[0-9]|3[0-1])/([0-9][0-9])");
    if (re.test(format)) {
        return month + '/' + day + '/' + year.slice(-2);
    }
    // 06-07-2008
    re = new RegExp("^(0[1-9]|1[0-2])-(0[1-9]|1[0-9]|2[0-9]|3[0-1])-([0-9][0-9][0-9][0-9])");
    if (re.test(format)) {
        return month + '-' + day + '-' + year;
    }
    // 06-07-08
    re = new RegExp("^(0[1-9]|1[0-2])-(0[1-9]|1[0-9]|2[0-9]|3[0-1])-([0-9][0-9])");
    if (re.test(format)) {
        return month + '-' + day + '-' + year.slice(-2);
    }
    // Saturday, Jun 7, 2008
    re = new RegExp("^(Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday),? Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec, (0?[1-9]|1[0-9]|2[0-9]|3[0-1]) ([0-9][0-9][0-9][0-9])", "i");
    if (re.test(format)) {
        let dayOfWeek = new Date(date).getDay();
        console.log(dayOfWeek);
        return daysFull[dayOfWeek] + ', ' + months[month - 1] + ' ' + Number(day) + ', ' + year;
    }
    // June 7, 2008
    re = new RegExp("^January|February|March|April|May|June|July|August|September|October|November|December, (0?[1-9]|1[0-9]|2[0-9]|3[0-1]) ([0-9][0-9][0-9][0-9])", "i");
    if (re.test(format)) {
        return monthsFull[month - 1] + ' ' + Number(day) + ', ' + year;
    }
    // Jun 7, 2008
    re = new RegExp("^Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec, (0?[1-9]|1[0-9]|2[0-9]|3[0-1]) ([0-9][0-9][0-9][0-9])", "i");
    if (re.test(format)) {
        return months[month - 1] + ' ' + Number(day) + ', ' + year;
    }
    // Can't find format
    return date;
}