**Build**: make build
**Run**: docker run --rm -v /Users/a/share:/share hellofresh -input_file=/share/fixtures.json

**Help**: make help - shows available flags

To run application fixtures file needs to be attached. To do that:
1. mount volume by using docker flag: -v /Users/a/share:/share
2. use application flag: -input_file=/share/fixtures.json, with path to file in mounted volume

Optional flags:
* **recipe_names**: Recipe names to search joined by comma or another delimiter if provided. Default: "".
* **recipe_names_delimiter**: Delimiter for recipe names to search. Default: ",".
* **deliveries_by_postcode**: Count deliveries in JSON file with postcode. Default: "10120".
* **deliveries_by_postcode_from_time**: Count deliveries in JSON file with postcode after provided time. Default: "11AM".
* **deliveries_by_postcode_to_time**: Count deliveries in JSON file with postcode until provided time. Default: "3PM".


Taken decisions:
1. File is big - that's why opened as stream and processed items one by one - didn't have to load all file in memory.
2. The most time is spent on file reading and this operation cannot be paralleled. 
Since the file is big, I decided to parallel file reading and data processing - didn't check the overall speed change, 
but the file reading will not be paused by data processing.
3. used uint32 as type for all integers - all variables are positive and will store lower values
