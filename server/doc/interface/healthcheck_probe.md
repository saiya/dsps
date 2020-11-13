# GET `/probe/liveness`

Liveness probe is an endpoint to check this server process should be killed or not.

## Request

Neither request parameter nor request body needed.

## Response

This endpoint returns HTTP `200` (OK) if healthy.

Body of response is only for investigation, do not programmatically rely on the body.



# GET `/probe/readiness`

Readiness probe is an endpoint to check this server process can accept requests or not.

## Request

Neither request parameter nor request body needed.

## Response

This endpoint returns HTTP `200` (OK) if healthy.

Body of response is only for investigation, do not programmatically rely on the body.
