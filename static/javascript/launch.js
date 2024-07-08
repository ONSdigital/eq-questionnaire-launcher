// uuidv4: from https://github.com/kelektiv/node-uuid
!(function (e) {
  if ("object" == typeof exports && "undefined" != typeof module)
    module.exports = e();
  else if ("function" == typeof define && define.amd) define([], e);
  else {
    var n;
    (n =
      "undefined" != typeof window
        ? window
        : "undefined" != typeof global
          ? global
          : "undefined" != typeof self
            ? self
            : this),
      (n.uuidv4 = e());
  }
})(function () {
  return (function e(n, r, o) {
    function t(f, u) {
      if (!r[f]) {
        if (!n[f]) {
          var a = "function" == typeof require && require;
          if (!u && a) return a(f, !0);
          if (i) return i(f, !0);
          var d = new Error("Cannot find module '" + f + "'");
          throw ((d.code = "MODULE_NOT_FOUND"), d);
        }
        var p = (r[f] = { exports: {} });
        n[f][0].call(
          p.exports,
          function (e) {
            var r = n[f][1][e];
            return t(r ? r : e);
          },
          p,
          p.exports,
          e,
          n,
          r,
          o,
        );
      }
      return r[f].exports;
    }
    for (
      var i = "function" == typeof require && require, f = 0;
      f < o.length;
      f++
    )
      t(o[f]);
    return t;
  })(
    {
      1: [
        function (e, n, r) {
          function o(e, n) {
            var r = n || 0,
              o = t;
            return [
              o[e[r++]],
              o[e[r++]],
              o[e[r++]],
              o[e[r++]],
              "-",
              o[e[r++]],
              o[e[r++]],
              "-",
              o[e[r++]],
              o[e[r++]],
              "-",
              o[e[r++]],
              o[e[r++]],
              "-",
              o[e[r++]],
              o[e[r++]],
              o[e[r++]],
              o[e[r++]],
              o[e[r++]],
              o[e[r++]],
            ].join("");
          }
          for (var t = [], i = 0; i < 256; ++i)
            t[i] = (i + 256).toString(16).substr(1);
          n.exports = o;
        },
        {},
      ],
      2: [
        function (e, n, r) {
          var o =
            ("undefined" != typeof crypto &&
              crypto.getRandomValues &&
              crypto.getRandomValues.bind(crypto)) ||
            ("undefined" != typeof msCrypto &&
              "function" == typeof window.msCrypto.getRandomValues &&
              msCrypto.getRandomValues.bind(msCrypto));
          if (o) {
            var t = new Uint8Array(16);
            n.exports = function () {
              return o(t), t;
            };
          } else {
            var i = new Array(16);
            n.exports = function () {
              for (var e, n = 0; n < 16; n++)
                0 === (3 & n) && (e = 4294967296 * Math.random()),
                  (i[n] = (e >>> ((3 & n) << 3)) & 255);
              return i;
            };
          }
        },
        {},
      ],
      3: [
        function (e, n, r) {
          function o(e, n, r) {
            var o = (n && r) || 0;
            "string" == typeof e &&
              ((n = "binary" === e ? new Array(16) : null), (e = null)),
              (e = e || {});
            var f = e.random || (e.rng || t)();
            if (((f[6] = (15 & f[6]) | 64), (f[8] = (63 & f[8]) | 128), n))
              for (var u = 0; u < 16; ++u) n[o + u] = f[u];
            return n || i(f);
          }
          var t = e("./lib/rng"),
            i = e("./lib/bytesToUuid");
          n.exports = o;
        },
        { "./lib/bytesToUuid": 1, "./lib/rng": 2 },
      ],
    },
    {},
    [3],
  )(3);
});

// store fetch so it only needs to be re-done if the survey changes
let supplementaryDataSets = null;

// We always need survey_id from top-level schema metadata for SDS retrieval
let schemaSurveyId = null;

function clearSurveyMetadataFields() {
  document
    .querySelector("#survey-type-metadata-accordion")
    .classList.add("ons-u-vh");
  document.querySelector("#survey_metadata_fields").innerHTML = "";
  setTabIndex("survey_type_metadata_detail", -1);
  showMetadataAccordion("sds", false);
}

function setSurveyType(event) {
  localStorage.setItem("survey_type", event.value);
  setLaunchType("remote");
}

function setCirSchema(event) {
  localStorage.setItem("cir_schema", event.value);
  setLaunchType("cir");
}

function setSchemaUrl(event) {
  localStorage.setItem("schema_url", event.value);
  setLaunchType("url");
}

function setLaunchType(launchType) {
  const schemaName = document.querySelector("#schema_name");
  const schemaUrl = document.querySelector("#schema-url");
  const cirSchemas = document.querySelector("#cir-schemas");
  console.log(schemaName);
  const remoteSchemaSurveyType = document.querySelector(
    "#remote-schema-survey-type",
  );

  if (["cir", "remote", "url"].includes(launchType)) {
    if (schemaName.selectedIndex) {
      clearSurveyMetadataFields();
      enableSubmitFlushButtons(false);
      schemaName.selectedIndex = 0;
      localStorage.removeItem("schema_name");
    }

    if (launchType === "cir") {
      schemaUrl.value = "";
      localStorage.removeItem("schema_url");
    } else if (launchType === "url") {
      cirSchemas.selectedIndex = 0;
      localStorage.removeItem("cir_schema");
    }
  }
  if (launchType === "name") {
    schemaUrl.value = "";
    cirSchemas.selectedIndex = 0;
    remoteSchemaSurveyType.selectedIndex = 0;
    localStorage.removeItem("schema_url");
    localStorage.removeItem("cir_schema");
    localStorage.removeItem("survey_type");
    document.querySelector("#language_code").disabled = false;
  }
}

function showMetadataAccordion(type, show) {
  let accordionElement = document.querySelector(`#${type}-metadata-accordion`);
  if (show) {
    accordionElement.classList.remove("ons-u-vh");
  } else {
    accordionElement.classList.add("ons-u-vh");
  }
}

function enableButtons(button) {
  for (let i = 0; i < button.length; i++) {
    button[i].classList.remove("ons-btn--disabled");
    button[i].disabled = false;
  }
}

function disableButtons(button) {
  for (let i = 0; i < button.length; i++) {
    button[i].classList.add("ons-btn--disabled");
    button[i].disabled = true;
  }
}

function enableSubmitFlushButtons(enable) {
  let submitButton = document.querySelector("#submit-btn");
  let flushButton = document.querySelector("#flush-btn");

  let buttons = [submitButton, flushButton];

  if (enable) {
    enableButtons(buttons);
  } else {
    disableButtons(buttons);
  }
}

function includeSurveyMetadataFields(schema_name, survey_type) {
  let formTypeValue = schema_name.split("_").slice(1).join("_");
  document
    .querySelector("#survey-type-metadata-accordion")
    .classList.remove("ons-u-vh");
  document.querySelector(".survey_heading").innerHTML =
    `${survey_type} Survey Metadata`;

  document.querySelector("#survey_metadata_fields").innerHTML =
    `<div class="ons-field ons-field--inline">
                    <label class="ons-label" for="form_type">form_type</label>
                    <input id="form_type" name="form_type" type="text" value="${formTypeValue}" class="ons-input ons-input--text ons-input-type__input">
                </div>`;
  setTabIndex("survey_type_metadata_detail", 0);
  showMetadataAccordion("sds", true);
}

function loadMetadataForSchemaName() {
  let schemaName = document.querySelector("#schema_name").value;
  localStorage.setItem("schema_name", schemaName);

  if (schemaName !== "Select Schema") {
    const surveyType = document.querySelector(
      `#schema_name option[value="${schemaName}"]`,
    ).dataset.surveyType;
    loadSurveyMetadata(schemaName, surveyType);
    loadSchemaMetadata(schemaName, null);
  }
}

function loadMetadataForRemoteSchema() {
  let schemaUrl = document.querySelector("#schema-url").value;
  let surveyType = document.querySelector("#remote-schema-survey-type");

  let cirSchemaDropdown = document.querySelector("#cir-schemas");
  let cirInstrumentId = cirSchemaDropdown.selectedIndex
    ? cirSchemaDropdown.value
    : null;

  let schemaName = null;

  if (schemaUrl && !schemaUrl.endsWith(".json")) {
    alert("Schema URL is not valid URL must end with '.json'");
    return false;
  }

  if (!surveyType.selectedIndex) {
    alert("Select a Survey Type.");
    return false;
  }

  if (!schemaUrl && !cirInstrumentId) {
    alert("Enter a Schema URL or select a CIR Schema.");
    return false;
  }

  if (schemaUrl) {
    schemaName = schemaUrl.split("/").slice(-1)[0].split(".json")[0];
    document.querySelector("#language_code").disabled = false;
  } else {
    let cirSchema = cirSchemaDropdown.options[cirSchemaDropdown.selectedIndex];
    schemaName = cirSchema.getAttribute("data-form-type");
    let language = cirSchema.getAttribute("data-language");

    showCIRMetdata(cirInstrumentId, cirSchema);

    // cir schemas are for a specific language, so populate and disable choosing it
    populateDropDownWithValue("#language_code", language);
    document.querySelector("#language_code").disabled = true;
  }

  loadSurveyMetadata(schemaName, surveyType.value);
  loadSchemaMetadata(schemaName, schemaUrl, cirInstrumentId);
  enableSubmitFlushButtons(true);
}

function loadSurveyMetadata(schema_name, survey_type) {
  if (
    survey_type.toLowerCase() === "test" ||
    survey_type.toLowerCase() === "social"
  ) {
    clearSurveyMetadataFields();
  } else {
    includeSurveyMetadataFields(schema_name, survey_type);
  }
}

async function getDataAsync(queryParam) {
  return new Promise((resolve, reject) => {
    let xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function () {
      if (this.readyState === 4) {
        if (this.status === 200) {
          resolve(JSON.parse(this.responseText));
        } else {
          alert(`Request failed. ${this.responseText}`);
          reject(`Request failed. ${this.responseText}`);
        }
      }
    };
    xhttp.open("GET", queryParam, true);
    xhttp.send();
  });
}

function getLabelFor(fieldName) {
  return `<label class="ons-label" for="${fieldName}">${fieldName}</label>`;
}

function getInputField(
  fieldName,
  type,
  defaultValue = null,
  isReadOnly = false,
  onChangeCallback = null,
) {
  const value = defaultValue ? `value="${defaultValue}"` : "";
  const readOnly = isReadOnly ? "readonly" : "";
  if (readOnly) {
    return `<input ${readOnly} id="${fieldName}" type="${type}" ${value} class="ons-input ons-input--text ons-input--w-20" onchange="${onChangeCallback}">`;
  }
  if (type === "checkbox") {
    return `<input ${readOnly} id="${fieldName}" type="${type}" ${value} class="ons-checkbox--toggle" onchange="${onChangeCallback}">`;
  }
  return `<input ${readOnly} id="${fieldName}" name="${fieldName}" type="${type}" ${value} class="ons-input ons-input--text ons-input--w-20" onchange="${onChangeCallback}">`;
}

async function loadSDSDatasetMetadata(survey_id, period_id) {
  if (survey_id && period_id) {
    const sds_dataset_metadata_url = `/supplementary-data?survey_id=${survey_id}&period_id=${period_id}`;
    return await getDataAsync(sds_dataset_metadata_url);
  }
  return null;
}

function handleNoSupplementaryData() {
  showMetadataAccordion("sds", false);
  setTabIndex("sds_metadata_detail", -1);
  enableSubmitFlushButtons(false);
}

function showCIRMetdata(cirInstrumentId, cirSchema) {
  showMetadataAccordion("cir", true);
  let ciMetadata = {
    id: cirInstrumentId,
    ci_version: cirSchema.getAttribute("data-version"),
    title: cirSchema.getAttribute("data-title"),
    description: cirSchema.getAttribute("data-description"),
  };
  document.querySelector("#cir_metadata").innerHTML = Object.keys(ciMetadata)
    .map(
      (key) =>
        `<div class="ons-field ons-field--inline">${getLabelFor(key)}${getInputField(key, "text", ciMetadata[key], true)}</div>`,
    )
    .join("");
  setTabIndex("cir_metadata_detail", 0);
}

function updateSDSDropdown() {
  const surveyId = schemaSurveyId;
  const periodId = document.getElementById("period_id")?.value;

  const supplementaryDataSection = document.querySelector("#sds_id");
  const sdsDatasetIdElement = document.querySelector("#sds_dataset_id");
  loadSDSDatasetMetadata(surveyId, periodId)
    .then((sds_metadata_response) => {
      if (sds_metadata_response?.length) {
        document.querySelector("#sds_id").innerHTML = "";
        supplementaryDataSets = sds_metadata_response;
        showMetadataAccordion("sds", true);
        setTabIndex("sds_metadata_detail", 0);
        enableSubmitFlushButtons(true);

        if (
          !document
            .querySelector("#survey_metadata")
            .contains(sdsDatasetIdElement)
        ) {
          // add sds_dataset_id field into the SDS metadata section if not already in survey metadata
          supplementaryDataSection.innerHTML = `<div class="ons-field ons-field--inline">${getLabelFor("sds_dataset_id")}<select id="sds_dataset_id" name="sds_dataset_id" class="ons-input ons-input--select ons-input--w-20" onchange="loadSupplementaryDataInfo()"></select></div>`;
        }

        document.querySelector("#sds_dataset_id").innerHTML =
          sds_metadata_response
            .map(
              (dataset) =>
                `<option value="${dataset.dataset_id}">${dataset.dataset_id}</option>`,
            )
            .join("");
        loadSupplementaryDataInfo();
      } else if (document.querySelector("#sds_dataset_id")) {
        document.querySelector("#sds_dataset_id").innerHTML = "";
        handleNoSupplementaryData();
      }
    })
    .catch((_) => {
      handleNoSupplementaryData();
    });
}

function loadSchemaMetadata(schemaName, schemaUrl, cirInstrumentId) {
  let survey_data_url = `/survey-data?`;

  if (cirInstrumentId) {
    survey_data_url += `&cir_instrument_id=${cirInstrumentId}`;
  } else {
    showMetadataAccordion("cir", false);
    setTabIndex("cir_metadata_detail", -1);
    if (schemaName) survey_data_url += `&schema_name=${schemaName}`;
    if (schemaUrl) survey_data_url += `&schema_url=${schemaUrl}`;
  }
  showMetadataAccordion("sds", false);
  setTabIndex("sds_metadata_detail", -1);
  getDataAsync(survey_data_url)
    .then((schema_response) => {
      document.querySelector("#survey_metadata").innerHTML = "";
      document.querySelector("#survey_metadata").innerHTML = "";

      // We always need survey_id from top-level schema metadata for SDS retrieval
      schemaSurveyId = schema_response.survey_id;

      if (schema_response.metadata.length > 0) {
        document.querySelector("#survey_metadata").innerHTML =
          schema_response.metadata
            .map((metadataField) => {
              const fieldName = metadataField["name"];
              const defaultValue = metadataField["default"];

              return `<div class="ons-field ons-field--inline">${getLabelFor(fieldName)}${(() => {
                if (metadataField["type"] === "boolean") {
                  return getInputField(fieldName, "checkbox");
                } else if (metadataField["type"] === "uuid") {
                  return (
                    `<span>${getInputField(fieldName, "text", uuidv4())}` +
                    `<img onclick="uuid('${fieldName}')" src="data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiA/PjwhRE9DVFlQRSBzdmcgIFBVQkxJQyAnLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4nICAnaHR0cDovL3d3dy53My5vcmcvR3JhcGhpY3MvU1ZHLzEuMS9EVEQvc3ZnMTEuZHRkJz48c3ZnIGhlaWdodD0iNTEycHgiIGlkPSJMYXllcl8xIiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCA1MTIgNTEyOyIgdmVyc2lvbj0iMS4xIiB2aWV3Qm94PSIwIDAgNTEyIDUxMiIgd2lkdGg9IjUxMnB4IiB4bWw6c3BhY2U9InByZXNlcnZlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIj48Zz48cGF0aCBkPSJNMjU2LDM4NC4xYy03MC43LDAtMTI4LTU3LjMtMTI4LTEyOC4xYzAtNzAuOCw1Ny4zLTEyOC4xLDEyOC0xMjguMVY4NGw5Niw2NGwtOTYsNTUuN3YtNTUuOCAgIGMtNTkuNiwwLTEwOC4xLDQ4LjUtMTA4LjEsMTA4LjFjMCw1OS42LDQ4LjUsMTA4LjEsMTA4LjEsMTA4LjFTMzY0LjEsMzE2LDM2NC4xLDI1NkgzODRDMzg0LDMyNywzMjYuNywzODQuMSwyNTYsMzg0LjF6Ii8+PC9nPjwvc3ZnPg==">` +
                    `</span>`
                  );
                } else if (
                  fieldName === "survey_id" ||
                  fieldName === "period_id"
                ) {
                  return getInputField(
                    fieldName,
                    "text",
                    fieldName === "survey_id"
                      ? schema_response.survey_id
                      : defaultValue,
                    false,
                    "updateSDSDropdown()",
                  );
                } else if (fieldName === "sds_dataset_id") {
                  return `<select id="${fieldName}" name="${fieldName}" class="ons-input ons-input--select ons-input--w-20" onchange="loadSupplementaryDataInfo()"></select>`;
                } else {
                  return getInputField(fieldName, "text", defaultValue);
                }
              })()}</div>`;
            })
            .join("");
        updateSDSDropdown();
      } else {
        document.querySelector("#survey_metadata").innerHTML =
          "No metadata required for this survey";
      }
      enableSubmitFlushButtons(true);
    })
    .catch((_) => {
      document.querySelector("#survey_metadata").innerHTML =
        "Failed to load Survey Metadata";
    });
}

function loadSupplementaryDataInfo() {
  const selectedDatasetId = document.getElementById("sds_dataset_id")?.value;
  const selectedDataset = supplementaryDataSets?.find(
    (d) => d["dataset_id"] === selectedDatasetId,
  );

  if (!selectedDataset) {
    return;
  }

  const sdsDatasetMetadataKeys = [
    "title",
    "total_reporting_units",
    "schema_version",
    "sds_dataset_version",
  ];

  const sdsMetadataSection = document.querySelector("#sds_data");
  const sdsMetadataField = (key) =>
    `<div class="ons-field ons-field--inline sds_data">${getLabelFor(key)}${getInputField(key, "text", selectedDataset[key], true)}</div>`;

  clearSDSMetadata();
  if (sdsMetadataSection.contains(document.querySelector("#sds_dataset_id"))) {
    sdsMetadataSection.innerHTML += sdsDatasetMetadataKeys
      .map(sdsMetadataField)
      .join("");
  } else {
    sdsMetadataSection.innerHTML = sdsDatasetMetadataKeys
      .map(sdsMetadataField)
      .join("");
  }
}

function clearSDSMetadata() {
  let dataset = document.querySelectorAll("#sds_data > div");
  let div_array = [...dataset];
  for (let i = 0; i < div_array.length; i++) {
    if (div_array[i].classList.contains("sds-data")) {
      div_array[i].remove();
    }
  }
}

function uuid(el_id) {
  document.querySelector(`#${el_id}`).value = uuidv4();
}

function numericId() {
  let result = "";
  let chars = "0123456789";
  for (let i = 16; i > 0; --i) {
    result += chars[Math.round(Math.random() * (chars.length - 1))];
  }
  document.querySelector(`#response_id`).value = result;
}

function setResponseExpiry(days_offset = 7) {
  let dt = new Date();
  dt.setDate(dt.getDate() + days_offset);
  document.querySelector("#response_expires_at").value = dt
    .toISOString()
    .replace(/(\.\d*)/, "")
    .replace(/Z/, "+00:00");
}

function validateForm() {
  validateResponseExpiresAt();
  removeUnwantedMetadata();
}

function validateResponseExpiresAt() {
  let responseExpiresAt = Date.parse(
    document.querySelector("#response_expires_at").value,
  );
  if (isNaN(responseExpiresAt)) {
    document.querySelector("#response_expires_at").remove();
  }
}

// Inputs without a name will not be submitted
function removeUnwantedMetadata() {
  const inputs = document.getElementsByTagName("input");
  for (let input of inputs) {
    if (!input.value) {
      input.removeAttribute("name");
    }
  }
}

function retrieveResponseId() {
  let responseId = localStorage.getItem("response_id");
  let responseIdButton = document.querySelector("#response-id-btn");

  if (responseId) {
    responseIdButton.classList.remove("ons-btn--disabled");
    responseIdButton.disabled = false;
  } else {
    responseIdButton.classList.add("ons-btn--disabled");
    responseIdButton.disabled = true;
  }
}

function loadResponseId() {
  document.querySelector("#response_id").value =
    localStorage.getItem("response_id");
}

function saveResponseId() {
  localStorage.setItem(
    "response_id",
    document.querySelector("#response_id").value,
  );
}

function clearLocalStorage() {
  localStorage.removeItem("response_id");
  localStorage.removeItem("schema_name");
  localStorage.removeItem("survey_type");
  localStorage.removeItem("cir_schema");
  localStorage.removeItem("schema_url");
  location.reload();
}

function populateDropDownWithValue(selector, value) {
  const availableOptions = [...document.querySelector(selector).options].map(
    (x) => x.value,
  );

  if (availableOptions.includes(value)) {
    document.querySelector(selector).value = value;
  }
}

function setTabIndex(metadataDetail, value) {
  document.getElementById(metadataDetail).tabIndex = value;
}

function initialiseTabIndex() {
  const details = [
    "cir_metadata_detail",
    "survey_type_metadata_detail",
    "sds_metadata_detail",
  ];
  for (i = 0; i < details.length; i++) {
    document.getElementById(details[i]).tabIndex = -1;
  }
}

function onLoad() {
  uuid("collection_exercise_sid");
  uuid("case_id");
  numericId();
  setResponseExpiry();
  retrieveResponseId();
  initialiseTabIndex();

  if ((schemaName = localStorage.getItem("schema_name"))) {
    populateDropDownWithValue("#schema_name", schemaName);
    loadMetadataForSchemaName();
  } else {
    if ((surveyType = localStorage.getItem("survey_type"))) {
      populateDropDownWithValue("#remote-schema-survey-type", surveyType);
    }
    if ((cirSchema = localStorage.getItem("cir_schema"))) {
      populateDropDownWithValue("#cir-schemas", cirSchema);
    }
    if ((schemaUrl = localStorage.getItem("schema_url"))) {
      document.querySelector("#schema-url").value = schemaUrl;
    }
  }
}
