var last_id = '';
var loading = false;

var $form = $('form#filters');

$('form select').change(function () {

    $.each($form.serializeArray(), function (index, value) {
        if (value.value === '') {
            $('select[name=' + value.name + ']').attr('disabled', 'disabled');
        }
    });

    $form.submit();
});

$(window).scroll(function () {
    if ($(window).scrollTop() + $(window).height() > $(document).height() - 3000) {
        getMore();
    }
});

getMore();

function getMore() {

    if (loading === false) {

        console.log('getting more');

        loading = true;

        $.ajax({
            method: "GET",
            url: "/listing",
            data: {
                last: last_id,
            },
            success: function (data, status, xhr) {

                if ("error" in data) {

                    $('ul#results').html($('<li>Please login</li>'));

                } else {

                    last_id = data.last_id;

                    var transform = {
                        '<>': 'li', 'class': 'media mb-1', 'data-id': '${id}', 'html': [
                            {
                                '<>': 'a', 'href': '${link}', 'html': [
                                    {'<>': 'img', 'class': 'mr-3', 'src': '${icon}', 'alt': '${title}', 'width': '140px;'},
                                ]
                            },
                            {
                                '<>': 'div', 'class': 'media-body', 'html': [
                                    {'<>': 'h5', 'class': 'mt-0 mb-1', 'html': '${title}'},
                                    {
                                        '<>': 'p', 'class': 'mb-0', 'html': [
                                            {'<>': 'a', 'href': '/r/${reddit}', 'html': '/r/${reddit}'}
                                        ]
                                    },
                                    {
                                        '<>': 'p', 'class': 'mb-0', 'html': [
                                            {'<>': 'a', 'href': '${comments_link}', 'html': '${comments_count} Comments'}
                                        ]
                                    }
                                ]
                            },
                        ]
                    };

                    $('.spinner').remove();

                    $('ul#results').json2html(data.items, transform);

                    $('ul#results').append($('<hr>'));

                    loading = false;

                }
            }
        });
    }
}
