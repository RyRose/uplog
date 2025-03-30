/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/static/**/*.{html,js}",
    "./web/app/*.{html,js}",
    "./cmd/**/*.templ",
    "./internal/**/*.templ",
  ],
  plugins: [require("daisyui")],
};
