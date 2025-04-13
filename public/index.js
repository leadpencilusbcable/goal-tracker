const goalTable = document.getElementById("goal-table");
const goalTableBody = goalTable.querySelector("tbody");
const submitButton = document.getElementById("submit-button");

function addRowToGoalTable(){
  const goalTableRow = document.createElement("tr");

  goalTableRow.innerHTML = `
    <td><input type="text" name="title" required/></td>
    <td><textarea rows="1" name="notes"/></textarea></td>
    <td><input type="date" name="due" required/></td>
    <td>
      <button class="minus-button" onclick="removeRowFromGoalTable(this)" type="button">&minus;</button>
    </td>`;

  goalTableBody.appendChild(goalTableRow);

  if(goalTableBody.children.length > 0){
    submitButton.removeAttribute("disabled");
  }
}

/**
 * @param {HTMLButtonElement} button - the button from which this is called
 */
function removeRowFromGoalTable(button){
  let parentEle = button.parentElement;

  while(parentEle !== null){
    if(parentEle.nodeName === "TR"){
      parentEle.remove();

      if(goalTableBody.children.length === 0){
        submitButton.setAttribute("disabled", "");
      }

      return;
    }

    parentEle = parentEle.parentElement;
  }
}

