import http from "k6/http";

let routes = [
  "CR-Fairmount",
  "CR-Fitchburg",
  "CR-Worcester",
  "CR-Franklin",
  "CR-Greenbush",
  "CR-Haverhill",
  "CR-Kingston",
  "CR-Lowell",
  "CR-Middleborough",
  "CR-Needham",
  "CR-Newburyport",
  "CR-Providence",
  "CR-Foxboro",
  "Boat-F4",
  "Boat-F1",
  "Boat-EastBoston",
  "121",
  "201",
  "202",
  "210",
  "245",
  "351",
  "354",
];

function shuffleArray(array) {
  for (let i = array.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [array[i], array[j]] = [array[j], array[i]];
  }
}

export default function () {
  shuffleArray(routes);
  routes.forEach((route) => {
    http.get(`http://localhost:4000/schedules/${route}`);
  });
}
