# Modules Endpoints

* [Add Module](#Add-Module)
* [Download Module](#Download-Module)
* [Delete Module](#Delete-Module)
* [List Modules](#List-Modules)

## Add Module

Add module.

### Request

```
POST /v1/modules/${TEAM}/${PROJECT}/names/${NAME}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Module name. |

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

## Download Module

Download module by version.

### Request

```
GET /v1/modules/${TEAM}/${PROJECT}/names/${NAME}/versions/${VERSION}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Module name. |
| **VERSION** | Module version. |

### Response

```
Data as a file.
```

<br/>

## Delete Module

Delete module by version.

### Request

```
DELETE /v1/modules/${TEAM}/${PROJECT}/names/${NAME}/versions/${VERSION}
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Module name. |
| **VERSION** | Module version. |

<br/>

## List Modules

List modules.

### Request

```
GET /v1/modules/list

GET /v1/modules/${TEAM}/${PROJECT}/list

GET /v1/modules/${TEAM}/${PROJECT}/names/${NAME}/list
```

#### Parameters

| Parameter | Description |
|:----------|:------------|
| **TEAM** | Team name. |
| **PROJECT** | Project name. |
| **NAME** | Module name. |

### Response

```
[
    {
        "created": <string>
        "id": <int>
        "name": <string>
        "project": <string>
        "team": <string>
        "updated": <string>
        "version": <int>
    }
]
```
