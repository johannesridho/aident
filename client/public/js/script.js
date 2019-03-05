$(document).ready(function () {
    $('form input').change(function (e) {
        var fileName = e.target.files[0].name;
        $('form p').text(fileName);
    });
});