const goalInputTable = document.getElementById("goal-input-table");
const goalInputTableBody = goalInputTable.querySelector("tbody");
const submitButton = document.getElementById("submit-button");
const goalDisplayLoadingSpinner = document.getElementsByClassName("loading-spinner")[0].cloneNode(false);

/**
 * @param {String} name
 */
function deleteCookie(name){
  document.cookie = name + "=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;"
}

loadGoalDisplayTable();

function getLocalDateString(){
  const date = new Date();

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

async function loadGoalDisplayTable(){
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const url =  "/goals?timezone=" + timezone;

  const res = await fetch(url);

  const container = document.getElementById("goal-display-container");

  if(res.status === 204) {
    container.innerHTML = "No goals.";
  } else {
    container.innerHTML = await res.text();
  }
}

async function refreshDisplayTable(){
  const container = document.getElementById("goal-display-container");
  container.replaceChildren(goalDisplayLoadingSpinner);

  loadGoalDisplayTable();
}

/**
 * @param {Event} event
 */
async function submitGoals(event){
  event.preventDefault();

  const form = event.target;
  const formData = new FormData(form);

  const today_date_str = getLocalDateString();

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

