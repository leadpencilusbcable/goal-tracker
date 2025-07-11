const goalInputTable = document.getElementById("goal-input-table");
const goalInputTableBody = goalInputTable.querySelector("tbody");
const submitButton = document.getElementById("submit-button");
const goalDisplayLoadingSpinner = document.getElementsByClassName("loading-spinner")[0].cloneNode(false);
const startFilter = document.getElementById("start-filter");
const endFilter = document.getElementById("end-filter");
const statusCheckboxes = document.getElementsByName("status");
const viewButton = document.getElementById("view-button");
const makeButton = document.getElementById("make-button");

/**
 * @param {Event} event
 */
function startFilterOnChange(event){
  goalParams.start = event.target.value;
  refreshDisplayTable();
}

/**
 * @param {Event} event
 */
function endFilterOnChange(event){
  goalParams.end = event.target.value;
  refreshDisplayTable();
}

/**
 * @param {"In progress" | "Complete" | "Failed"} status
 */
function statusFilterOnChange(status){
  let checkedCount = 0;

  for(let i = 0; i < 3; i++){
    if(goalParams.statuses[i]){
      checkedCount++;
    }
  }

  let statusIndex = 0;

  if(status == "Complete") statusIndex = 1;
  else if(status == "Failed") statusIndex = 2;

  //if there is only 1 checkbox checked, don't allow uncheck
  if(checkedCount === 1 && goalParams.statuses[statusIndex]){
    statusCheckboxes[statusIndex].checked = true;
    return;
  }

  goalParams.statuses[statusIndex] = !goalParams.statuses[statusIndex];
  refreshDisplayTable();
}

/**
 * @param {"view" | "make"} view
 */
function switchView(view){
  const goalForm = document.getElementById("goal-form");
  const goalView = document.getElementById("goal-view");

  switch(view){
    case "view":
      viewButton.setAttribute("disabled", "");
      makeButton.removeAttribute("disabled");

      goalForm.hidden = true;
      goalView.hidden = false;

      break;
    case "make":
      makeButton.setAttribute("disabled", "");
      viewButton.removeAttribute("disabled");

      goalView.hidden = true;
      goalForm.hidden = false;

      break;
  }
}

/**
 * @typedef GoalParams
 * @type {object}
 * @property {Date} start
 * @property {Date} end
 * @property {Boolean[]} statuses
 */

/** @type {GoalParams} */
let goalParams;

function init(){
  viewButton.setAttribute("disabled", "");

  const now = new Date();
  const nowPlusOneWeek = addDaysToDate(now, 7);

  const start = getLocalDateString(now);
  const end = getLocalDateString(nowPlusOneWeek);

  startFilter.value = start;
  endFilter.value = end;

  const statuses = [false, false, false];

  for(let i = 0; i < statusCheckboxes.length; i++){
    if(statusCheckboxes[i].checked){
      statuses[i] = true;
    }
  }

  goalParams = {
    start,
    end,
    statuses,
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
        <td style='text-align: center;'>
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
  startFilter.setAttribute("disabled", "");
  endFilter.setAttribute("disabled", "");

  for(let i = 0; i < statusCheckboxes.length; i++){
    statusCheckboxes[i].setAttribute("disabled", "");
  }

  const now = new Date();

  let url =
    "/goals?start=" +
    goalParams.start +
    "&end=" +
    goalParams.end +
    "&now=" +
    getLocalDateString(now);

  if(goalParams.statuses[0]){
    url += "&status=In progress";
  }
  if(goalParams.statuses[1]){
    url += "&status=Complete";
  }
  if(goalParams.statuses[2]){
    url += "&status=Failed";
  }

  const loadingSpinner = document.getElementsByClassName("loading-spinner")[0];
  const res = await fetch(url);

  if(res.status === 204){
    loadingSpinner.outerHTML = "<p class='no-goal-text'>No goals</p>";
  } else {
    loadingSpinner.outerHTML = await res.text();
  }

  startFilter.removeAttribute("disabled");
  endFilter.removeAttribute("disabled");

  for(let i = 0; i < statusCheckboxes.length; i++){
    statusCheckboxes[i].removeAttribute("disabled");
  }
}

function refreshDisplayTable(){
  const container = document.getElementById("display-container");
  container.innerHTML = '';
  container.appendChild(goalDisplayLoadingSpinner);

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
    <td style='text-align: center;'>
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

