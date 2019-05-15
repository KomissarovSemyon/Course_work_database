// contries_en
// https://en.wikipedia.org/wiki/ISO_3166-1
console.log(JSON.stringify(
    Array.from(table.querySelectorAll('tr'))
        .filter(tr => tr.querySelector('td'))
        .map(tr => [
            tr.querySelectorAll('td')[1].innerText,
            tr.querySelectorAll('td')[0].innerText.trim()
        ])
))

// contries_ru
// https://ru.wikipedia.org/wiki/ISO_3166-1
console.log(JSON.stringify(
    Array.from(table.querySelectorAll('tr'))
        .filter(tr => tr.querySelector('span[class="nowrap"]'))
        .map(tr => [
            tr.querySelectorAll('td')[1].innerText,
            tr.querySelector('span[class="nowrap"]').innerText.trim()
        ])
))