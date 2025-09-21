function footerf() {
    const
        main = document.getElementsByTagName('main')[0],
        footer = document.getElementsByTagName('footer')[0],
        susick = document.getElementById('mobile-nav');
    main.style.paddingBottom = footer.clientHeight-susick.clientHeight + 'px';
}

window.addEventListener('load', footerf);
window.addEventListener('resize', footerf);