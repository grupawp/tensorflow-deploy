# Models Endpoints

* [Add Model](#Add-Model)
* [Download Model](#Download-Model)
* [Set Model Label](#Set-Model-Label)
* [Revert Stable Label](#Revert-Stable-Label)
* [Delete Model Label](#Delete-Model-Label)
* [Delete Model](#Delete-Model)
* [List Models](#List-Models)
* [Reload Models](#Reload-Models)
* [Get TFS Config](#Get-TFS-Config)

## Add Model

Add model with optional label.

### Request

```
POST /v1/models/${TEAM}/${PROJECT}/names/${NAME}

POST /v1/models/${TEAM}/${PROJECT}/names/${NAME}/labels/${LABEL}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |
| **LABEL** | Label name which will be assinged to the model. |

### Response

```
{
    "name": <string>
    "project": <string>
    "team": <string>
    "version": <int>
}
```

<br/>

## Download Model

Download model by version or label.

### Request

```
GET /v1/models/${TEAM}/${PROJECT}/names/${NAME}/versions/${VERSION}

GET /v1/models/${TEAM}/${PROJECT}/names/${NAME}/labels/${LABEL}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |
| **VERSION** | Model version. |
| **LABEL** | Label name assigned to the model. |

### Response

```
Data as a file.
```

<br/>

## Set Model Label
Set label of the model.

### Request

```
PUT /v1/models/${TEAM}/${PROJECT}/names/${NAME}/versions/${VERSION}/labels/stable

PUT /v1/models/${TEAM}/${PROJECT}/names/${NAME}/versions/${VERSION}/labels/${LABEL}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |
| **VERSION** | Model version. |
| **LABEL** | Label name which will be assinged to the model. |

<br/>

## Revert Stable Label

Revert `stable` label to the previous `stable` model version.

### Request

```
PUT /v1/models/${TEAM}/${PROJECT}/names/${NAME}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |

<br/>

## Delete Model Label

Delete label assigned to the model.

### Request

```
DELETE /v1/models/${TEAM}/${PROJECT}/names/${NAME}/lables/${LABEL}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |
| **LABEL** | Label name. |

<br/>

## Delete Model

Delete model by version or label.

### Request

```
DELETE /v1/models/${TEAM}/${PROJECT}/names/${NAME}/versions/${VERSION}

DELETE /v1/models/${TEAM}/${PROJECT}/names/${NAME}/lables/${LABEL}/remove_version
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |
| **VERSION** | Model version. |
| **LABEL** | Label name. |

<br/>

## List Models

List models.

### Request

```
GET /v1/models/list

GET /v1/models/${TEAM}/${PROJECT}/list

GET /v1/models/${TEAM}/${PROJECT}/names/${NAME}/list
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Model name. |

### Response

```
[
    {
        "created": <string>
        "id": <int>
        "label": <string>
        "name": <string>
        "project": <string>
        "status": <string>
        "team": <string>
        "updated": <string>
        "version": <int>
    }
]
```

<br/>

## Reload Models

Reload models within team-project.

### Request

```
POST /v1/models/${TEAM}/${PROJECT}/reload
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |

<br/>

## Get TFS Config

Get TFS models configuration within team-project.

### Request

```
GET /v1/models/${TEAM}/${PROJECT}/config
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |


### Response

```
Data as a file.
```

#### Example
```
model_config_list: <
  config: <
    name: "name"
    base_path: "/models/team/project/name"
    model_platform: "tensorflow"
    model_version_policy: <
      specific: <
        versions: 1
        versions: 2
        versions: 3
      >
    >
    version_labels: <
      key: "canary"
      value: 3
    >
    version_labels: <
      key: "label"
      value: 2
    >
  >
>
```
