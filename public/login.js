/**
 * @param {Event} event
 */
async function handleLogin(event){
  event.preventDefault();

  const form = event.target;
  const formData = new FormData(form);
  const urlFormData = new URLSearchParams(formData);

  const res = await fetch(form.action, {
    body: urlFormData,
    method: form.method,
  });

  const res_body = await res.text();

  if(res.ok){
    window.location.href = "/";
  } else {
    document.getElementById("login-error").style.display = "flex";
    document.getElementById("login-error-text").innerText = res_body;
  }
}
