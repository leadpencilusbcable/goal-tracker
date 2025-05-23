const goalInputTable = document.getElementById("goal-input-table");
const goalInputTableBody = goalInputTable.querySelector("tbody");
const submitButton = document.getElementById("submit-button");
const goalDisplayLoadingSpinner = document.getElementsByClassName("loading-spinner")[0].cloneNode(false);

/**
 * @typedef GoalParams
 * @type {object}
 * @property {Date} start
 * @property {Date} end
 */

/** @type {GoalParams} */
let goalParams;

function init(){
  const now = new Date();
  const nowPlusOneWeek = addDaysToDate(now, 7);

  goalParams = {
    start: now,
    end: nowPlusOneWeek,
  };

  loadGoalDisplayTable(goalParams);
}

init();

/**
 * @param {String} name
 */
function deleteCookie(name){
  document.cookie = name + "=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;"
}


/**
 * @param {Date} date
 */
function getLocalDateString(date){
  const day = date.getDate();
  const month = date.getMonth();
  const year = date.getFullYear();

  const str = (
    year + '-' +
    (month + 1).toString().padStart(2, '0') + '-' +
    day.toString().padStart(2, '0')
  );

  return str;
}

/**
 * @param {Date} date
 * @param {Number} days
 */
function addDaysToDate(date, days){
  const newDate = new Date(date);
  newDate.setDate(date.getDate() + days);
  return newDate;
}

async function logout(){
  await fetch("/logout", { method: "POST" });

  deleteCookie("session_id");
  window.location.replace("/login");
}

function resetGoalInputTable(){
  for(let i = 0; i < goalInputTableBody.children.length; i++){
    const child = goalInputTableBody.children[i];

    if (i === 0){
      child.innerHTML = `
        <td><input type="text" name="title" required/></td>
        <td><textarea rows="1" name="notes"/></textarea></td>
        <td><input type="date" name="due" required/></td>
        <td>
          <button class="minus-button" disabled onclick="removeRowFromGoalTable(this)" type="button">&minus;</button>
        </td>`;
    } else {
      child.remove();
    }
  }
}

/**
 * @param {GoalParams} goalParams
 */
async function loadGoalDisplayTable(goalParams){
  const now = new Date();

  const url =
    "/goals?start=" +
    getLocalDateString(goalParams.start) +
    "&end=" +
    getLocalDateString(goalParams.end) +
    "&now=" +
    getLocalDateString(now);

  const res = await fetch(url);

  const container = document.getElementById("goal-display-container");

  container.innerHTML = await res.text();
}

async function refreshDisplayTable(){
  const container = document.getElementById("goal-display-container");
  container.replaceChildren(goalDisplayLoadingSpinner);

  loadGoalDisplayTable(goalParams);
}

/**
 * @param {Event} event
 */
async function submitGoals(event){
  event.preventDefault();

  const form = event.target;
  const formData = new FormData(form);

  const today_date_str = getLocalDateString(new Date());

  for(let i = 0; i < formData.getAll("title").length; i++){
    formData.append("start", today_date_str);
  }
  const urlFormData = new URLSearchParams(formData);

  const res = await fetch(form.action, {
    body: urlFormData,
    method: form.method,
  });

  const res_body = await res.text();

  if(res.ok){
    //TODO set success message
    form.reset();
    resetGoalInputTable();
    refreshDisplayTable();
  } else {
    alert(res_body);
  }
}

function addRowToGoalTable(){
  if(goalInputTableBody.children.length === 1) {
    goalInputTableBody.querySelector(".minus-button").removeAttribute("disabled");
  }

  const goalTableRow = document.createElement("tr");

  goalTableRow.innerHTML = `
    <td><input type="text" name="title" required/></td>
    <td><textarea rows="1" name="notes"/></textarea></td>
    <td><input type="date" name="due" required/></td>
    <td>
      <button class="minus-button" onclick="removeRowFromGoalTable(this)" type="button">&minus;</button>
    </td>`;

  goalInputTableBody.appendChild(goalTableRow);
}

/**
 * @param {HTMLButtonElement} button - the button from which this is called
 */
function removeRowFromGoalTable(button){
  let parentEle = button.parentElement;

  while(parentEle !== null){
    if(parentEle.nodeName === "TR"){
      parentEle.remove();

      if(goalInputTableBody.children.length === 1) {
        goalInputTableBody.querySelector(".minus-button").setAttribute("disabled", "");
      }

      return;
    }

    parentEle = parentEle.parentElement;
  }
}

