function openModal() {
    document.getElementById('loginModal').style.cssText = 'opacity: 1; visibility: visible;';
}

function closeModal() {
    document.getElementById('loginModal').style.cssText = 'opacity: 0; visibility: hidden';
}

document.getElementById('loginModal').addEventListener('click', function(t) {
    if (t.target === this) closeModal();
});