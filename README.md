# MoCo MD Public Safety App 

The purpose of this application is to make it easy for prospective home-buyers to easily determine whether or not a neighborhood is safe in the Montgomery County, MD area. 

This application consumes crime incident reports from the publically available MoCo police department public safety data set and uses basic geospatial functions to display crime incidents at the voting precinct level.

Crime incident reports are processed, counted, and stored in a Spatialite database. 

This application is written in Go and utilizes the Orb Geospatial library.

The resultant dataset is displayed below in QGIS. The darker the blue, the more crimes were committed in the precinct.
