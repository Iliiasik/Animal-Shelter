function filterAnimals(species) {
    console.log('Filtering animals by species:', species); // Для отладки
    const url = new URL(window.location.href);
    url.searchParams.set('species', species);
    console.log('Redirecting to URL:', url.toString()); // Для отладки
    window.location.href = url.toString();
}
