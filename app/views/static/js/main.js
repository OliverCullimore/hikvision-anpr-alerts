/* Cookie */
const cookie = {
    set: function(name, value, days) {
        let expires = "";
        if (days) {
            let date = new Date();
            date.setTime(date.getTime() + (days*24*60*60*1000));
            expires = "; expires=" + date.toUTCString();
        }
        document.cookie = name + "=" + (value || "")  + expires + "; path=/";
    },
    get: function(name) {
        let nameEQ = name + "=";
        let ca = document.cookie.split(';');
        for (let i=0;i < ca.length;i++) {
            let c = ca[i];
            while (c.charAt(0)==' ') c = c.substring(1,c.length);
            if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
        }
        return null;
    }
};

/* Theme */
const theme = {
    current: cookie.get('theme') ?? 'light',
    update: function() {
        cookie.set('theme', this.current, 300);
        document.documentElement.setAttribute('data-theme', this.current);
    },
    toggle: function() {
        this.current = (this.current === 'dark' ? 'light' : 'dark');
        this.update();
    }
};
theme.update();
window.addEventListener("focus", theme.update());
document.getElementById('theme-toggle').addEventListener('click', function() {
    theme.toggle();
});