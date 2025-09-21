function close_button_function(event) {
    const el = event.target;
    let parent = el.parentNode;
    while !(parent.classList.contains(class)){
        parent = parent.parentNode;
    }
    parent.classList.remove("show");
}