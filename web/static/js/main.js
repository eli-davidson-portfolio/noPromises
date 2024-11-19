function showNewFlowModal() {
    document.getElementById('new-flow-modal').style.display = 'block';
}

function hideNewFlowModal() {
    document.getElementById('new-flow-modal').style.display = 'none';
}

// Close modal when clicking outside
window.onclick = function(event) {
    const modal = document.getElementById('new-flow-modal');
    if (event.target === modal) {
        hideNewFlowModal();
    }
}

// Initialize HTMX handlers
document.addEventListener('DOMContentLoaded', function() {
    // Load initial flow list
    htmx.ajax('GET', '/api/v1/flows', {target: '#flow-list'});
}); 