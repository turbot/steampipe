
steampipe service start
for (( c=1; c<=100; c++ ))
do

  echo file 1
  cp -f ~/.steampipe/config/src/aws1.spc  ~/.steampipe/config/aws.spc
  arn1=$(steampipe query "select distinct arn from aws_g1.aws_account" --output json | jq -cs '.[0][0].arn')
  arn2=$(steampipe query "select distinct arn from aws_g2.aws_account" --output json | jq -cs '.[0][0].arn')
  arn3=$(steampipe query "select distinct arn from aws_g3.aws_account" --output json | jq -cs '.[0][0].arn')
  arn4=$(steampipe query "select distinct arn from aws_g4.aws_account" --output json | jq -cs '.[0][0].arn')

  if [ "$arn1" = "\"arn:aws:::876515858155"\" ] &&  [ "$arn2" = "\"arn:aws:::533793682495"\" ] &&  [ "$arn3" = "\"arn:aws:::097350876455"\" ] &&  [ "$arn4" = "\"arn:aws:::882789663776"\" ]
  then
    echo "OK"
  else
    echo "BAD"
  fi
  sleep 5

  echo file 2
  cp -f ~/.steampipe/config/src/aws2.spc  ~/.steampipe/config/aws.spc

   arn1=$(steampipe query "select distinct arn from aws_g1.aws_account" --output json | jq -cs '.[0][0].arn')
   arn2=$(steampipe query "select distinct arn from aws_g2.aws_account" --output json | jq -cs '.[0][0].arn')
   arn3=$(steampipe query "select distinct arn from aws_g3.aws_account" --output json | jq -cs '.[0][0].arn')
   arn4=$(steampipe query "select distinct arn from aws_g4.aws_account" --output json | jq -cs '.[0][0].arn')

   if [ "$arn1" = "\"arn:aws:::882789663776"\" ] && [ "$arn2" = "\"arn:aws:::876515858155"\" ] &&  [ "$arn3" = "\"arn:aws:::533793682495"\" ] &&  [ "$arn4" = "\"arn:aws:::097350876455"\" ]
   then
     echo "OK"
   else
     echo "BAD"
     c=100
   fi

   sleep 5

done

steampipe service stop

