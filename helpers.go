package main

func checkError(err error) {
	if err != nil {
		errorLogger.Fatal(err)
	}
}
