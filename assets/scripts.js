var last_id = '';
var loading = false;
var stop = false;
var page = 1;

var $form = $('form#filters');
var $results = $('ul#results');
var $loadButton = $('#load-more');

$loadButton.click(function () {
    getMore();
});

$form.find('select').change(function () {

    $.each($form.serializeArray(), function (index, value) {
        if (value.value === '') {
            $form.find('[name=' + value.name + ']').attr('disabled', 'disabled');
        }
    });

    $form.submit();
});

$(window).scroll(function () {
    if ($(window).scrollTop() + $(window).height() > $(document).height() - 5000) {
        getMore();
    }
});

getMore();

function getMore() {

    if (loading === false && stop !== true) {

        start_loading();

        $.ajax({
            method: "GET",
            url: "/ajax" + window.location.search,
            data: {
                last: last_id,
                reddit: reddit,
            },
            success: function (data, status, xhr) {

                if ("error" in data) {

                    $results.append($('<li>' + data.error + '</li>'));
                    stop = true;
                    $loadButton.remove();

                } else if (data.items !== null) {

                    last_id = data.last_id;

                    var transform = {
                        '<>': 'li', 'class': 'media mb-1', 'data-id': '${id}', 'html': [
                            {
                                '<>': 'a', 'target': '_blank', 'href': '${link}', 'html': [
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
                                            {'<>': 'a', 'target': '_blank', 'href': '${comments_link}', 'html': '${comments_count} Comments'}
                                        ]
                                    }
                                ]
                            },
                        ]
                    };

                    $results.json2html(data.items, transform);
                    $results.append($('<div class="card page-number mb-1">Page ' + page + '</div>'));

                } else {

                    // No error and no items
                    last_id = data.last_id;
                    $results.append($('<div class="card page-number mb-1">Page ' + page + '</div>'));

                }

                page++;

                // sleep_ms(1000);

                stop_loading();
            }
        });
    }
}

function start_loading() {

    console.log('Loading...');
    loading = true;
    $loadButton.attr('disabled', 'disabled');
    $loadButton.find('i').show();

}

function stop_loading() {

    console.log('Complete.');
    loading = false;
    $loadButton.removeAttr('disabled');
    $loadButton.find('i').hide();

}

function sleep_ms(millisecs) {
    var initiation = new Date().getTime();
    while ((new Date().getTime() - initiation) < millisecs) ;
}
