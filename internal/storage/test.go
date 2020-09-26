package storage

import mgr "github.com/t-bfame/diago/internal/manager"

func AddTest(test *mgr.Test) error {
	return daoFactory.GetTestDao().AddTest(test)
}

func GetTestByTestId(testId mgr.TestID) (*mgr.Test, error) {
	return daoFactory.GetTestDao().GetTestByTestId(testId)
}
