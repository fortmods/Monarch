document.querySelector("[data-action='launch-game']").onclick = () => {
    try {
        const username = document.querySelector("[data-content='username']").value;
        if (!username) return alert("Enter a username then try again");

        window.go.main.App.Run(username).catch(err => console.error(err));
    } catch (err) {
        console.error(err);
    };
};

document.querySelector("[data-action='close-game']").onclick = () => {
    try {
        window.go.main.App.Close().catch(err => console.error(err));
    } catch (err) {
        console.error(err);
    };
};