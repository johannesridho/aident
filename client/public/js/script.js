$(document).ready(function () {
    $('form input').change(function (e) {
        var fileName = e.target.files[0].name;
        $('form p').text(this.files.length + fileName);
    });
});