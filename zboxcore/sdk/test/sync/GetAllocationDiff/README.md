# Structure definition of blobbers_response.json

JSON object definition for mocking blobber's HTTP response:

```
[
  {
    "method": "GET",
    "path": "/v1/file/list/7827c760363c4836b0acd3967a023c9061deef50e0c44db3e3aca8e14f8ef6f8",
    "params": [
      {
        "path_hash": "7a3507d80cbe8977dd602de3b7411d820b75c26f5d1d7da06e97276e2a63fc30"
      },
      {
        "path_hash": "706cbafb54abbc029213498322ddf996361cc3712ce8a2bd0b66b9c85a6055ac"
      }
    ],
    "responses": [
    [
        // according to params[0] condition, the response for each blobbers would be as bellow
        {"blobber_1_mock_response_for_path_hash_param1_as_json_object"},
        {"blobber_2_mock_response_for_path_hash_param1_as_json_object"},
        {"blobber_3_mock_response_for_path_hash_param1_as_json_object"},
        {"blobber_4_mock_response_for_path_hash_param1_as_json_object"},
        ...
    ],
    [
        // according to params[2] condition, the response for each blobbers would be as bellow
        {"blobber_1_mock_response_for_path_hash_param2_as_json_object"},
        {"blobber_2_mock_response_for_path_hash_param2_as_json_object"},
        {"blobber_3_mock_response_for_path_hash_param2_as_json_object"},
        {"blobber_4_mock_response_for_path_hash_param2_as_json_object"},
        ...
    ],
...
}
```
