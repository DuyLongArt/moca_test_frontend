# Handling error
When error occurs when the function can not function properly, write error with http.Error (w, "Can not generate token", error http can not handle server error)


# General respose hander

The response should has:
+ data
+ status
+ code
+ message

w.Header().set("Content-Type", "application/json")

w.WriteHeader(http.StatusOK)

w.Write([] byte {`"status":"OK","data:","code", "message":`})
