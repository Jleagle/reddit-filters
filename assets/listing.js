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

// Save buttons
$('#results').on('click', 'p[data-saved] i', function () {

    var $star = $(this);
    var id = $star.closest('li').attr('data-id');

    var $parent = $star.parent();
    var saved = !!+$parent.attr('data-saved');

    console.log(saved);

    $(this).attr('class', 'fas fa-spinner fa-spin');

    $.ajax({
        method: "GET",
        url: "/ajax/" + (saved ? "unsave" : "save"),
        data: {
            id: id,
        },
        success: function (data, status, xhr) {
            if (data === 'OK') {
                if (saved) {
                    $parent.attr('data-saved', '0');
                    $star.attr('class', 'far fa-star');
                } else {
                    $parent.attr('data-saved', '1');
                    $star.attr('class', 'fas fa-star');
                }
            }
        }
    });
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
            url: "/ajax/listing" + window.location.search,
            data: {
                last: last_id,
                reddit: reddit,
            },
            success: function (data, status, xhr) {

                if ("error" in data) {

                    $results.append($('<li>' + data.error + '</li>'));

                    stop = true;
                    $loadButton.remove();

                    stop_loading();

                } else if (data.items !== null) {

                    last_id = data.last_id;

                    var transform = {
                        '<>': 'li', 'class': 'media mb-1', 'data-id': '${id}', 'html': [
                            {
                                '<>': 'a', 'target': '_blank', 'href': '${link}', 'html': [
                                    {
                                        '<>': 'img', 'class': 'mr-3', 'src': '${icon}', 'width': '140px;', 'onerror': function (e) {
                                            $(this).attr('src', '/assets/logo.png');
                                        }
                                    },
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
                                    },
                                    {
                                        '<>': 'p', 'html': function () {
                                            return this.saved ? '<i class="fas fa-star"></i>' : '<i class="far fa-star"></i>'
                                        }, 'data-saved': function () {
                                            return this.saved ? '1' : '0'
                                        }
                                    }
                                ]
                            },
                        ]
                    };

                    $results.json2html(data.items, transform);
                    $results.append($('<div class="card page-number mb-1">Page ' + page + '</div>'));

                    page++;

                    stop_loading();

                } else {

                    last_id = data.last_id;

                    $results.append($('<div class="card page-number mb-1">Page ' + page + '</div>'));

                    page++;

                    stop_loading();

                    getMore();

                }
            }
        });
    }
}

function start_loading() {

    loading = true;
    $loadButton.attr('disabled', 'disabled');
    $loadButton.find('i').show();

}

function stop_loading() {

    loading = false;
    $loadButton.removeAttr('disabled');
    $loadButton.find('i').hide();

}
