function openModal(m) {
    document.getElementById(m).style.cssText = 'opacity: 1; visibility: visible;';
}

function closeModal(m) {
    document.getElementById(m).style.cssText = 'opacity: 0; visibility: hidden';
}

document.getElementById('loginModal').addEventListener('click', function(t) {
    if (t.target == this) closeModal('loginModal');
});

document.getElementById('beginModal').addEventListener('click', function(t) {
    if (t.target == this) closeModal('beginModal');
});