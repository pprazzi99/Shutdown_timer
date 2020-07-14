import time as t 
import os
from win10toast import ToastNotifier #delete if you don't have win10toast installed!!!

notify = ToastNotifier() #delete if you don't have win10toast installed!!!

#! variables
hours = 0
minutes = 0
seconds = 0
x = 0

#! functions
def zero():
    hours = 0
    minutes = 0
    seconds = 0

def check():
    if minutes > 60:
        print("Wrong amount!!!")
        t.sleep(1)
        x = 2
        zero()
    elif seconds > 60:
        print("Wrong amount!!!")
        t.sleep(1)
        x = 2
        zero()
    else:
        return 1

#! main
while x != 1:
    os.system('cls')
    print("Remember to switch off quick_edit!!!")
    print("In what time do you want your PC to be switched off?")
    try: 
        hours = int(input("Hours: "))
        x = check()
    except ValueError:
        print("Wrong value")
        t.sleep(1)
        zero()
        x = 2
    
    if x != 2:
        try:
            minutes = int(input("Minutes: "))
            x = check()
        except ValueError:
            print("Wrong value")
            t.sleep(1)
            zero()
            x = 2
        
    if x != 2:
        if x == 1: 
            try:
                seconds = int(input("Seconds: "))
                x = check()
            except ValueError:
                print("Wrong value")
                t.sleep(1)
                zero()
                x = 2
            
    
while seconds != -1:
    if hours > 0:
        if minutes > 0:
            if seconds > 0:
                seconds -=1
                t.sleep(1)
            else: 
                minutes -=1
                seconds = 60
        elif minutes == 0:
            if seconds > 0:
                seconds -=1
                t.sleep(1)
            else:
                hours -=1
                minutes = 60
    elif hours == 0 and minutes > 0:
        if seconds > 0:
                seconds -=1
                t.sleep(1)
        else: 
            minutes -=1
            seconds = 60
    elif hours == 0 and minutes == 0 and seconds >= 0:
        seconds -=1
        t.sleep(1)
    
    if seconds >= 0: 
        os.system('cls')
        print("Time left: "+str(hours)+":"+str(minutes)+":"+str(seconds))
        print("\nTo stop press CTRL+C")

    if hours == 0 and minutes == 5 and seconds == 0:
        notify.show_toast(title="5 minutes to shutdown left", msg="Timer", threaded=True, duration=240) #delete if you don't have win10toast installed!!!
    
    if hours == 0 and minutes == 0 and seconds == 30:
        notify.show_toast(title="30 seconds to shutdown left!", msg="Timer", threaded=True, duration=30) #delete if you don't have win10toast installed!!!

    if hours == 0 and minutes == 0 and seconds == 0:
        os.system('shutdown /s /t 1')
