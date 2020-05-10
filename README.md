<h2 align="center">
  <a href="https://github.com/iamseki?tab=repositories">
    <img alt="Open Weather Logo" src="https://s3.amazonaws.com/media-p.slid.es/uploads/383894/images/1810601/a-3.png" width="250px" />
  </a>
</h2>
<h2 align="center">
  List services concurrently in GO 
</h2>

<p align="center">This program list and write in <strong>json</strong> format a file.log RDS and Elasticache services available for specific region and IAM role, using go routines.</p>
 <p align="center">A <strong>go routine</strong> is a function that is capable of running concurrently with
other functions.</p>

<p align="center">
  <a href="https://github.com/iamseki">
    <img alt="Made by Christian Seki" src="https://img.shields.io/badge/made%20by-Christian%20Seki-brightgreen">
  </a>
  <img alt="License" src="https://img.shields.io/badge/license-MIT-%2304D361">
</p>

---
### run with Build 

> execute `go build main.go` 

*then* `region=us-east-1 role=prod ./main` 

-  region can be any *aws* region options such as: `sa-east-1` , `us-east-1` and so on.
-  role must be `prod` , `stage` or `old`
---